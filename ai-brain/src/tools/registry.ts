/**
 * Tool Registry - manages all available tools for the AI Brain
 */

import { Tool } from "./tool";
import { scanTool, pocTool, crawlTool, reportTool, suggestTemplatesTool } from "./built_in";

export class ToolRegistry {
  private tools: Map<string, Tool> = new Map();

  constructor() {
    this.registerBuiltIns();
  }

  private registerBuiltIns() {
    this.register(scanTool);
    this.register(pocTool);
    this.register(crawlTool);
    this.register(reportTool);
    this.register(suggestTemplatesTool);
  }

  register(tool: Tool) {
    this.tools.set(tool.name, tool);
  }

  get(name: string): Tool | undefined {
    return this.tools.get(name);
  }

  list(): Tool[] {
    return Array.from(this.tools.values());
  }

  listNames(): string[] {
    return Array.from(this.tools.keys());
  }

  async call(name: string, args: any): Promise<any> {
    const tool = this.tools.get(name);
    if (!tool) {
      throw new Error(`Tool not found: ${name}`);
    }
    return await tool.call(args);
  }

  getToolDefinitionsForLLM(): any[] {
    return this.list().map((t) => ({
      name: t.name,
      description: t.description,
      input_schema: t.inputSchema,
    }));
  }
}
