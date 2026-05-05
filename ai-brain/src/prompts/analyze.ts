/**
 * Target Analysis Prompt
 * Analyzes target information and generates attack surface assessment
 */

export interface AnalyzeParams {
  target: string;
  open_ports?: number[];
  services?: Array<{ port: number; product?: string; version?: string }>;
  vulnerabilities?: string[];
}

export function analyzeTargetPrompt(params: AnalyzeParams): string {
  const servicesText =
    params.services
      ?.map(
        (s) =>
          `- Port ${s.port}: ${s.product || "unknown"}${s.version ? " " + s.version : ""}`
      )
      .join("\n") || "No service information available";

  const vulnsText =
    params.vulnerabilities?.map((v) => `- ${v}`).join("\n") ||
    "No known vulnerabilities";

  return `Analyze the following target for offensive security assessment:

TARGET: ${params.target}
OPEN PORTS: ${params.open_ports?.join(", ") || "unknown"}

SERVICES:
${servicesText}

KNOWN VULNERABILITIES:
${vulnsText}

Provide your analysis as a JSON object with this exact structure:
{
  "risk_score": <number 0-10>,
  "attack_surface": ["<vector1>", "<vector2>"],
  "recommended_next_steps": ["<action1>", "<action2>"],
  "high_value_targets": ["<target1>", "<target2>"],
  "reasoning": "<detailed reasoning>"
}

Guidelines:
- Risk score: 10 = trivial remote root, 0 = no attack surface
- Attack surface: list specific attack vectors (e.g., "unauthenticated Redis", "Apache Struts RCE", "default credentials")
- Next steps: prioritize by expected ROI (information gain vs effort)
- High value targets: services that would yield maximum impact if compromised`;
}
