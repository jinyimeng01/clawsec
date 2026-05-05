/**
 * Tool system for AI Brain
 * Inspired by claude-code-haha's modular Tool architecture
 */

export interface ToolInputSchema {
  type: "object";
  properties: Record<string, { type: string; description: string; enum?: string[] }>;
  required?: string[];
}

export interface Tool {
  name: string;
  description: string;
  inputSchema: ToolInputSchema;
  isReadOnly: boolean;
  isDestructive: boolean;
  call: (args: any) => Promise<any> | any;
}

export function defineTool(config: {
  name: string;
  description: string;
  inputSchema: ToolInputSchema;
  isReadOnly?: boolean;
  isDestructive?: boolean;
  call: (args: any) => Promise<any> | any;
}): Tool {
  return {
    name: config.name,
    description: config.description,
    inputSchema: config.inputSchema,
    isReadOnly: config.isReadOnly ?? true,
    isDestructive: config.isDestructive ?? false,
    call: config.call,
  };
}
