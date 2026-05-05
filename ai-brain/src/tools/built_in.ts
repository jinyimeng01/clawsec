/**
 * Built-in tools that map to clawsec core capabilities
 */

import { defineTool } from "./tool";

export const scanTool = defineTool({
  name: "scan",
  description: "Run a network port scan against a target. Returns open ports and service banners.",
  inputSchema: {
    type: "object",
    properties: {
      target: { type: "string", description: "Target IP, CIDR, or hostname" },
      ports: { type: "string", description: "Port range: top100, top1000, or custom like 80,443,8080-8090" },
      banner: { type: "boolean", description: "Enable banner grabbing" },
    },
    required: ["target"],
  },
  isReadOnly: true,
  isDestructive: false,
  call: async (args: { target: string; ports?: string; banner?: boolean }) => {
    // Simulation: in a real setup, this would spawn clawsec scan subprocess
    return {
      simulated: true,
      tool: "clawsec scan",
      command: `clawsec scan port -t ${args.target} -p ${args.ports || "top100"}${args.banner ? " --banner" : ""}`,
      note: "This is a tool definition. The Go core will execute the actual scan.",
      example_result: {
        open_ports: [22, 80, 443, 3306],
        services: [
          { port: 22, product: "OpenSSH", version: "8.2p1" },
          { port: 80, product: "Apache", version: "2.4.41" },
        ],
      },
    };
  },
});

export const pocTool = defineTool({
  name: "poc",
  description: "Run PoC (Proof of Concept) vulnerability verification templates against targets.",
  inputSchema: {
    type: "object",
    properties: {
      target: { type: "string", description: "Target URL or host" },
      template: { type: "string", description: "Specific template ID or path" },
      severity: { type: "string", description: "Filter by severity: critical, high, medium, low, info" },
    },
    required: ["target"],
  },
  isReadOnly: true,
  isDestructive: false,
  call: async (args: { target: string; template?: string; severity?: string }) => {
    return {
      simulated: true,
      tool: "clawsec poc",
      command: `clawsec poc run -u ${args.target}${args.template ? " -t " + args.template : ""}${args.severity ? " --severity " + args.severity : ""}`,
      note: "PoC execution requires --authorized flag.",
    };
  },
});

export const crawlTool = defineTool({
  name: "crawl",
  description: "Enumerate directories and files on a web target.",
  inputSchema: {
    type: "object",
    properties: {
      target: { type: "string", description: "Target URL" },
      wordlist: { type: "string", description: "Path to wordlist file (optional)" },
      ext: { type: "boolean", description: "Enable smart extensions" },
    },
    required: ["target"],
  },
  isReadOnly: true,
  isDestructive: false,
  call: async (args: { target: string; wordlist?: string; ext?: boolean }) => {
    return {
      simulated: true,
      tool: "clawsec crawl",
      command: `clawsec crawl dir -t ${args.target}${args.ext ? " --ext" : ""}`,
      note: "Directory enumeration is read-only.",
    };
  },
});

export const reportTool = defineTool({
  name: "report",
  description: "Generate a penetration test report from findings.",
  inputSchema: {
    type: "object",
    properties: {
      target: { type: "string", description: "Target name" },
      format: { type: "string", description: "Report format: html, markdown, json", enum: ["html", "markdown", "json"] },
      findings: { type: "string", description: "JSON array of findings" },
    },
    required: ["target"],
  },
  isReadOnly: true,
  isDestructive: false,
  call: async (args: { target: string; format?: string; findings?: string }) => {
    return {
      simulated: true,
      tool: "clawsec ai report",
      command: `clawsec ai report -t ${args.target} -o report.${args.format || "md"}`,
      note: "Report generation is safe to run anytime.",
    };
  },
});

export const suggestTemplatesTool = defineTool({
  name: "suggest_templates",
  description: "Suggest relevant PoC templates based on service fingerprints.",
  inputSchema: {
    type: "object",
    properties: {
      fingerprint: { type: "string", description: "Service fingerprint string, e.g. 'Apache/2.4.41, PHP/7.4'" },
      target: { type: "string", description: "Target URL" },
    },
    required: ["fingerprint"],
  },
  isReadOnly: true,
  isDestructive: false,
  call: async (args: { fingerprint: string; target?: string }) => {
    return {
      simulated: true,
      tool: "clawsec ai suggest",
      command: `clawsec ai suggest -t ${args.target || ""} -f "${args.fingerprint}"`,
    };
  },
});
