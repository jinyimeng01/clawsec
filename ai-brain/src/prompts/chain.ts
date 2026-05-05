/**
 * Exploit Chain Building Prompt
 */

export interface ChainParams {
  target: string;
  findings: Array<{
    type: string;
    detail: string;
    severity: string;
  }>;
  objective?: string;
}

export function buildChainPrompt(params: ChainParams): string {
  const findingsText = params.findings
    .map(
      (f, i) =>
        `${i + 1}. [${f.severity.toUpperCase()}] ${f.type}: ${f.detail}`
    )
    .join("\n");

  const objective = params.objective || "achieve maximum impact (RCE or sensitive data access)";

  return `You are an advanced red team operator. Build an exploit chain to ${objective}.

TARGET: ${params.target}

DISCOVERED FINDINGS:
${findingsText}

Respond with a JSON object:
{
  "steps": [
    {
      "step": 1,
      "action": "<what to do>",
      "tool": "<recommended tool>",
      "expected_outcome": "<what success looks like>",
      "fallback": "<what to do if this step fails>"
    }
  ],
  "overall_confidence": <0.0-1.0>
}

Guidelines:
- Each step should build on previous steps
- Include specific tools (e.g., "nuclei", "sqlmap", "metasploit")
- Fallbacks are mandatory
- Be honest about limitations`;
}
