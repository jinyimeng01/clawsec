/**
 * AI Security Brain - Core decision engine
 * Handles target analysis, PoC suggestion, exploit chain building, report generation
 */

import Anthropic from "@anthropic-ai/sdk";
import { z } from "zod";
import { MCPClient } from "./mcp/client";
import { ToolRegistry } from "./tools/registry";
import { analyzeTargetPrompt } from "./prompts/analyze";
import { suggestPoCsPrompt } from "./prompts/suggest";
import { buildChainPrompt } from "./prompts/chain";
import { generateReportPrompt } from "./prompts/report";

const AnalyzeResultSchema = z.object({
  risk_score: z.number().min(0).max(10),
  attack_surface: z.array(z.string()),
  recommended_next_steps: z.array(z.string()),
  high_value_targets: z.array(z.string()),
  reasoning: z.string(),
});

const SuggestResultSchema = z.object({
  templates: z.array(
    z.object({
      id: z.string(),
      name: z.string(),
      severity: z.string(),
      confidence: z.number(),
      reason: z.string(),
    })
  ),
  priority_order: z.array(z.string()),
});

const ChainResultSchema = z.object({
  steps: z.array(
    z.object({
      step: z.number(),
      action: z.string(),
      tool: z.string(),
      expected_outcome: z.string(),
      fallback: z.string(),
    })
  ),
  overall_confidence: z.number(),
});

export interface BrainConfig {
  apiKey: string;
  model?: string;
  baseURL?: string;
  mcpServers?: Array<{ name: string; command: string; args?: string[] }>;
}

export class Brain {
  private client: Anthropic;
  private model: string;
  private mcp: MCPClient;
  private tools: ToolRegistry;
  private conversationHistory: Array<{ role: "user" | "assistant"; content: string }> = [];

  constructor(config: BrainConfig) {
    this.client = new Anthropic({
      apiKey: config.apiKey,
      baseURL: config.baseURL,
    });
    this.model = config.model || "claude-sonnet-4-20250514";
    this.mcp = new MCPClient(config.mcpServers || []);
    this.tools = new ToolRegistry();
  }

  async initialize() {
    await this.mcp.connectAll();
  }

  // ==================== Tool System ====================

  listTools() {
    return this.tools.getToolDefinitionsForLLM();
  }

  async callTool(name: string, args: any) {
    return this.tools.call(name, args);
  }

  // ==================== Core Intelligence Methods ====================

  async analyzeTarget(params: {
    target: string;
    open_ports?: number[];
    services?: Array<{ port: number; product?: string; version?: string }>;
    vulnerabilities?: string[];
  }) {
    const prompt = analyzeTargetPrompt(params);
    const response = await this.callLLM(prompt);

    try {
      const json = this.extractJSON(response);
      return AnalyzeResultSchema.parse(json);
    } catch {
      return {
        risk_score: 5,
        attack_surface: ["unknown"],
        recommended_next_steps: ["run port scan", "identify services"],
        high_value_targets: [params.target],
        reasoning: response,
      };
    }
  }

  async suggestPoCs(params: {
    target: string;
    fingerprint: string;
    technologies?: string[];
  }) {
    const prompt = suggestPoCsPrompt(params);
    const response = await this.callLLM(prompt);

    try {
      const json = this.extractJSON(response);
      return SuggestResultSchema.parse(json);
    } catch {
      return {
        templates: [],
        priority_order: [],
      };
    }
  }

  async buildChain(params: {
    target: string;
    findings: Array<{
      type: string;
      detail: string;
      severity: string;
    }>;
    objective?: string;
  }) {
    const prompt = buildChainPrompt(params);
    const response = await this.callLLM(prompt);

    try {
      const json = this.extractJSON(response);
      return ChainResultSchema.parse(json);
    } catch {
      return {
        steps: [],
        overall_confidence: 0,
      };
    }
  }

  async generateReport(params: {
    target: string;
    findings: any[];
    format?: "markdown" | "html" | "json";
  }) {
    const prompt = generateReportPrompt(params);
    const response = await this.callLLM(prompt);
    return { report: response, format: params.format || "markdown" };
  }

  async chat(params: { message: string; context?: any }) {
    this.conversationHistory.push({ role: "user", content: params.message });

    const systemPrompt = `You are ClawSec AI, an elite offensive security expert. You help with:
- Penetration testing strategy and techniques
- Vulnerability analysis and exploitation
- Tool selection and command generation
- Security research and threat intelligence

Always think step by step. When suggesting commands, explain the rationale.
If the user asks for something illegal or unethical, refuse and explain why.`;

    const messages = this.conversationHistory.map((m) => ({
      role: m.role as "user" | "assistant",
      content: m.content,
    }));

    const response = await this.client.messages.create({
      model: this.model,
      max_tokens: 4096,
      system: systemPrompt,
      messages,
    });

    const text = response.content
      .filter((c) => c.type === "text")
      .map((c) => (c as any).text)
      .join("");

    this.conversationHistory.push({ role: "assistant", content: text });

    // Keep history manageable
    if (this.conversationHistory.length > 20) {
      this.conversationHistory = this.conversationHistory.slice(-10);
    }

    return { response: text };
  }

  async executeTool(params: { tool: string; args: any }) {
    // Try built-in tools first, then MCP
    const builtIn = this.tools.get(params.tool);
    if (builtIn) {
      return builtIn.call(params.args);
    }
    return this.mcp.execute(params.tool, params.args);
  }

  // ==================== Private Helpers ====================

  private async callLLM(prompt: string): Promise<string> {
    const response = await this.client.messages.create({
      model: this.model,
      max_tokens: 4096,
      system:
        "You are ClawSec AI Brain, an elite offensive security expert. Respond with structured JSON when possible.",
      messages: [{ role: "user", content: prompt }],
    });

    return response.content
      .filter((c) => c.type === "text")
      .map((c) => (c as any).text)
      .join("");
  }

  private extractJSON(text: string): any {
    // Try to find JSON in markdown code blocks
    const codeBlockMatch = text.match(/```(?:json)?\s*([\s\S]*?)```/);
    if (codeBlockMatch) {
      return JSON.parse(codeBlockMatch[1].trim());
    }
    // Try to find raw JSON object
    const jsonMatch = text.match(/\{[\s\S]*\}/);
    if (jsonMatch) {
      return JSON.parse(jsonMatch[0]);
    }
    throw new Error("No JSON found in response");
  }
}
