/**
 * ClawSec AI Brain - Entry Point
 * JSON-RPC server over stdio communicating with Go core
 */

import { Brain } from "./brain";
import { RPCServer } from "./rpc";

async function main() {
  const apiKey = process.env.ANTHROPIC_API_KEY || "";
  if (!apiKey) {
    console.error("[AI Brain] WARNING: ANTHROPIC_API_KEY not set. AI functions will fail.");
  }

  const config = {
    apiKey,
    model: process.env.AI_MODEL || "claude-sonnet-4-20250514",
    baseURL: process.env.AI_BASE_URL,
    mcpServers: parseMCPServers(),
  };

  const brain = new Brain(config);
  await brain.initialize();

  const server = new RPCServer(brain);
  server.start();

  console.error("[AI Brain] Ready. Waiting for JSON-RPC requests...");
}

function parseMCPServers(): Array<{ name: string; command: string; args?: string[] }> {
  try {
    const env = process.env.MCP_SERVERS;
    if (!env) return [];
    return JSON.parse(env);
  } catch {
    return [];
  }
}

main().catch((err) => {
  console.error("[AI Brain] Fatal error:", err);
  process.exit(1);
});
