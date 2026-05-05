/**
 * JSON-RPC 2.0 Server over stdio
 * Receives requests from Go core, routes to brain handlers
 */

import type { Brain } from "./brain";

export interface JSONRPCRequest {
  jsonrpc: "2.0";
  method: string;
  params?: any;
  id?: string | number | null;
}

export interface JSONRPCResponse {
  jsonrpc: "2.0";
  result?: any;
  error?: {
    code: number;
    message: string;
    data?: any;
  };
  id?: string | number | null;
}

export class RPCServer {
  private brain: Brain;
  private buffer = "";

  constructor(brain: Brain) {
    this.brain = brain;
  }

  start() {
    process.stdin.setEncoding("utf8");
    process.stdin.on("data", (chunk) => this.onData(chunk as string));
    process.stdin.on("end", () => {
      process.exit(0);
    });
  }

  private onData(chunk: string) {
    this.buffer += chunk;
    let lines = this.buffer.split("\n");
    this.buffer = lines.pop() || "";
    for (const line of lines) {
      if (line.trim()) {
        this.handleLine(line.trim());
      }
    }
  }

  private async handleLine(line: string) {
    let req: JSONRPCRequest;
    try {
      req = JSON.parse(line);
    } catch {
      this.sendError(null, -32700, "Parse error");
      return;
    }

    if (req.jsonrpc !== "2.0") {
      this.sendError(req.id, -32600, "Invalid Request");
      return;
    }

    try {
      const result = await this.route(req.method, req.params || {});
      this.sendResponse(req.id, result);
    } catch (err: any) {
      this.sendError(req.id, -32000, err.message || "Internal error");
    }
  }

  private async route(method: string, params: any): Promise<any> {
    switch (method) {
      case "ping":
        return { status: "ok", timestamp: Date.now() };

      case "analyze":
        return this.brain.analyzeTarget(params);

      case "suggest":
        return this.brain.suggestPoCs(params);

      case "chain":
        return this.brain.buildChain(params);

      case "report":
        return this.brain.generateReport(params);

      case "chat":
        return this.brain.chat(params);

      case "execute_tool":
        return this.brain.executeTool(params);

      default:
        throw new Error(`Unknown method: ${method}`);
    }
  }

  private sendResponse(id: string | number | null | undefined, result: any) {
    const resp: JSONRPCResponse = { jsonrpc: "2.0", result, id: id ?? null };
    process.stdout.write(JSON.stringify(resp) + "\n");
  }

  private sendError(
    id: string | number | null | undefined,
    code: number,
    message: string
  ) {
    const resp: JSONRPCResponse = {
      jsonrpc: "2.0",
      error: { code, message },
      id: id ?? null,
    };
    process.stdout.write(JSON.stringify(resp) + "\n");
  }
}
