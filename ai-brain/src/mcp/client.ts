/**
 * MCP (Model Context Protocol) Client
 * Connects to external tools and exposes them to the AI Brain
 */

import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";

interface MCPServerConfig {
  name: string;
  command: string;
  args?: string[];
}

export class MCPClient {
  private servers: MCPServerConfig[];
  private clients: Map<string, Client> = new Map();
  private transports: Map<string, StdioClientTransport> = new Map();

  constructor(servers: MCPServerConfig[]) {
    this.servers = servers;
  }

  async connectAll() {
    for (const server of this.servers) {
      try {
        await this.connect(server);
      } catch (err: any) {
        console.error(`[MCP] Failed to connect ${server.name}: ${err.message}`);
      }
    }
  }

  async connect(config: MCPServerConfig) {
    const transport = new StdioClientTransport({
      command: config.command,
      args: config.args || [],
    });

    const client = new Client(
      { name: "clawsec-ai", version: "0.1.0" },
      { capabilities: { tools: {}, resources: {} } }
    );

    await client.connect(transport);
    this.clients.set(config.name, client);
    this.transports.set(config.name, transport);

    console.error(`[MCP] Connected to ${config.name}`);
  }

  async execute(toolName: string, args: any): Promise<any> {
    for (const [name, client] of this.clients) {
      try {
        const tools = await client.listTools();
        const tool = tools.tools.find((t: any) => t.name === toolName);
        if (tool) {
          const result = await client.callTool({ name: toolName, arguments: args });
          return { server: name, result };
        }
      } catch {
        // Try next server
      }
    }
    return this.simulateTool(toolName, args);
  }

  async listTools(): Promise<string[]> {
    const allTools: string[] = [];
    for (const [name, client] of this.clients) {
      try {
        const tools = await client.listTools();
        allTools.push(...tools.tools.map((t: any) => `${name}/${t.name}`));
      } catch {
        // Ignore
      }
    }
    return allTools;
  }

  private simulateTool(toolName: string, args: any): any {
    switch (toolName) {
      case "nmap_scan":
        return {
          simulated: true,
          tool: "nmap",
          command: `nmap -sV ${args.target}`,
          note: "MCP server not connected. Install nmap MCP server for real execution.",
        };
      case "hydra_brute":
        return {
          simulated: true,
          tool: "hydra",
          command: `hydra -l ${args.username} -P ${args.wordlist} ${args.target} ${args.protocol}`,
          note: "MCP server not connected. Install hydra MCP server for real execution.",
        };
      case "sqlmap_scan":
        return {
          simulated: true,
          tool: "sqlmap",
          command: `sqlmap -u "${args.url}" --batch`,
          note: "MCP server not connected. Install sqlmap MCP server for real execution.",
        };
      default:
        return {
          error: `Tool ${toolName} not available. Connect an MCP server that provides this tool.`,
        };
    }
  }

  async disconnectAll() {
    for (const [, client] of this.clients) {
      await client.close();
    }
    this.clients.clear();
    this.transports.clear();
  }
}
