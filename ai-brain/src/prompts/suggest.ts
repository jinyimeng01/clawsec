/**
 * PoC Suggestion Prompt
 * Recommends vulnerability verification templates based on target fingerprint
 */

export interface SuggestParams {
  target: string;
  fingerprint: string;
  technologies?: string[];
}

export function suggestPoCsPrompt(params: SuggestParams): string {
  const techText =
    params.technologies?.join(", ") || params.fingerprint;

  return `You are a vulnerability research expert. Based on the target fingerprint, suggest the most relevant PoC templates to verify.

TARGET: ${params.target}
FINGERPRINT: ${params.fingerprint}
TECHNOLOGIES: ${techText}

Respond with a JSON object in this exact structure:
{
  "templates": [
    {
      "id": "<template-id>",
      "name": "<human readable name>",
      "severity": "critical|high|medium|low",
      "confidence": <0.0-1.0>,
      "reason": "<why this template is relevant>"
    }
  ],
  "priority_order": ["<template-id-1>", "<template-id-2>"]
}

Guidelines:
- Only suggest PoCs that are directly relevant to the detected technologies/versions
- Confidence should reflect how certain you are that this vulnerability exists (higher for version-specific CVEs)
- Prioritize: RCE > SQLi > Auth Bypass > Info Disclosure
- Include recent CVEs (2021-2026) for detected software versions
- If version is unknown, suggest generic detection templates with lower confidence`;
}
