/**
 * Report Generation Prompt
 * Generates professional penetration test reports
 */

export interface ReportParams {
  target: string;
  findings: any[];
  format?: "markdown" | "html" | "json";
}

export function generateReportPrompt(params: ReportParams): string {
  const findingsText = JSON.stringify(params.findings, null, 2);

  return `Generate a professional penetration test report.

TARGET: ${params.target}
DATE: ${new Date().toISOString().split("T")[0]}

FINDINGS:
${findingsText}

FORMAT: ${params.format || "markdown"}

Requirements:
1. Executive Summary (risk overview in business terms)
2. Technical Findings (detailed vulnerability descriptions)
3. Evidence (requests/responses, screenshots placeholders)
4. Impact Assessment (CVSS-style scoring with justification)
5. Remediation (prioritized, actionable fixes)
6. Attack Path Narrative (how an attacker could chain findings)

Style:
- Professional but readable by both technical and non-technical stakeholders
- Use severity colors: Critical (red), High (orange), Medium (yellow), Low (blue), Info (gray)
- Include specific command examples where relevant`;
}
