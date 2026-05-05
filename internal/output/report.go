package output

import (
	"fmt"
	"html"
	"strings"
	"time"
)

// ReportGenerator generates penetration test reports
type ReportGenerator struct {
	format string
}

// NewReportGenerator creates a report generator
func NewReportGenerator(format string) *ReportGenerator {
	return &ReportGenerator{format: format}
}

// Generate generates a report from findings
func (r *ReportGenerator) Generate(findings []Result, target string) string {
	switch r.format {
	case "html":
		return r.generateHTML(findings, target)
	case "markdown":
		return r.generateMarkdown(findings, target)
	case "json":
		return r.generateJSON(findings, target)
	default:
		return r.generateMarkdown(findings, target)
	}
}

func (r *ReportGenerator) generateMarkdown(findings []Result, target string) string {
	var sb strings.Builder

	sb.WriteString("# Penetration Test Report\n\n")
	sb.WriteString(fmt.Sprintf("**Target:** %s\n\n", target))
	sb.WriteString(fmt.Sprintf("**Date:** %s\n\n", time.Now().Format("2006-01-02")))
	sb.WriteString("---\n\n")

	// Executive Summary
	sb.WriteString("## Executive Summary\n\n")
	severityCounts := make(map[string]int)
	for _, f := range findings {
		severityCounts[f.Severity]++
	}

	if len(findings) == 0 {
		sb.WriteString("No vulnerabilities were identified during this assessment.\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("A total of **%d** findings were identified:\n\n", len(findings)))
		for sev, count := range severityCounts {
			sb.WriteString(fmt.Sprintf("- **%s**: %d\n", strings.ToUpper(sev), count))
		}
		sb.WriteString("\n")
	}

	// Findings
	sb.WriteString("## Technical Findings\n\n")
	for i, f := range findings {
		severityEmoji := getSeverityEmoji(f.Severity)
		sb.WriteString(fmt.Sprintf("### %s Finding %d: %s\n\n", severityEmoji, i+1, f.Name))
		sb.WriteString(fmt.Sprintf("- **Severity:** %s\n", f.Severity))
		sb.WriteString(fmt.Sprintf("- **Type:** %s\n", f.Type))
		sb.WriteString(fmt.Sprintf("- **Host:** %s\n", f.Host))
		if f.URL != "" {
			sb.WriteString(fmt.Sprintf("- **URL:** %s\n", f.URL))
		}
		sb.WriteString(fmt.Sprintf("- **Description:** %s\n\n", f.Message))

		if len(f.Extractor) > 0 {
			sb.WriteString("**Extracted Data:**\n\n")
			for k, v := range f.Extractor {
				sb.WriteString(fmt.Sprintf("- `%s`: %v\n", k, v))
			}
			sb.WriteString("\n")
		}

		sb.WriteString("---\n\n")
	}

	// Remediation
	sb.WriteString("## Remediation Recommendations\n\n")
	for i, f := range findings {
		remediation := getRemediation(f.Severity, f.Type)
		sb.WriteString(fmt.Sprintf("%d. **%s**: %s\n", i+1, f.Name, remediation))
	}
	sb.WriteString("\n")

	return sb.String()
}

func (r *ReportGenerator) generateHTML(findings []Result, target string) string {
	return generateDarkHTML(findings, target)
}

// generateDarkHTML creates a single-file, zero-dependency dark-themed report.
// Inspired by clawgod's security-themed web design.
func generateDarkHTML(findings []Result, target string) string {
	var sb strings.Builder

	severityCounts := make(map[string]int)
	for _, f := range findings {
		severityCounts[f.Severity]++
	}

	// Severity order for charts
	severityOrder := []string{"critical", "high", "medium", "low", "info"}
	severityColors := map[string]string{
		"critical": "#ff4444",
		"high":     "#ff8800",
		"medium":   "#ffcc00",
		"low":      "#00ccff",
		"info":     "#888888",
	}

	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>ClawSec Report - ` + html.EscapeString(target) + `</title>
<style>
:root{--bg:#0a0a0f;--surface:#12121a;--surface-2:#1a1a24;--border:#252535;--text:#e0e0e0;--text-dim:#888;--accent:#22c55e;--accent-dim:#16a34a;--critical:#ff4444;--high:#ff8800;--medium:#ffcc00;--low:#00ccff;--info:#888}
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:'Segoe UI',system-ui,-apple-system,sans-serif;background:var(--bg);color:var(--text);line-height:1.6;min-height:100vh}
.container{max-width:1100px;margin:0 auto;padding:24px}
header{background:linear-gradient(135deg,#0f172a 0%,#1e293b 100%);border:1px solid var(--border);border-radius:12px;padding:32px;margin-bottom:24px;position:relative;overflow:hidden}
header::before{content:'';position:absolute;top:0;right:0;width:300px;height:300px;background:radial-gradient(circle,rgba(34,197,94,0.08) 0%,transparent 70%);pointer-events:none}
.brand{display:flex;align-items:center;gap:12px;margin-bottom:16px}
.brand-icon{width:36px;height:36px;background:linear-gradient(135deg,var(--accent),var(--accent-dim));border-radius:8px;display:flex;align-items:center;justify-content:center;font-weight:700;font-size:18px;color:#000}
h1{font-size:28px;font-weight:700;margin-bottom:8px}
.meta{color:var(--text-dim);font-size:14px}
.summary-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:16px;margin-bottom:24px}
.card{background:var(--surface);border:1px solid var(--border);border-radius:10px;padding:20px;transition:transform .2s,border-color .2s}
.card:hover{transform:translateY(-2px);border-color:var(--accent-dim)}
.card h2{font-size:13px;text-transform:uppercase;letter-spacing:1px;color:var(--text-dim);margin-bottom:12px}
.stat-number{font-size:36px;font-weight:700;color:var(--accent)}
.sev-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(140px,1fr));gap:12px;margin-top:12px}
.sev-item{background:var(--surface-2);border-radius:8px;padding:14px;text-align:center;border-left:3px solid transparent}
.sev-item.critical{border-color:var(--critical)}.sev-item.high{border-color:var(--high)}.sev-item.medium{border-color:var(--medium)}.sev-item.low{border-color:var(--low)}.sev-item.info{border-color:var(--info)}
.sev-count{font-size:24px;font-weight:700}
.sev-label{font-size:11px;text-transform:uppercase;color:var(--text-dim);margin-top:4px}
.findings h2{font-size:20px;margin-bottom:20px;padding-bottom:12px;border-bottom:1px solid var(--border)}
.finding{background:var(--surface);border:1px solid var(--border);border-radius:10px;padding:24px;margin-bottom:16px;position:relative;overflow:hidden}
.finding::before{content:'';position:absolute;left:0;top:0;bottom:0;width:4px;background:var(--info)}
.finding.critical::before{background:var(--critical)}.finding.high::before{background:var(--high)}.finding.medium::before{background:var(--medium)}.finding.low::before{background:var(--low)}
.finding-header{display:flex;align-items:flex-start;justify-content:space-between;gap:16px;margin-bottom:12px;flex-wrap:wrap}
.finding-title{font-size:17px;font-weight:600}
.badge{display:inline-block;padding:4px 12px;border-radius:6px;font-size:11px;font-weight:700;text-transform:uppercase;letter-spacing:.5px}
.badge-critical{background:rgba(255,68,68,0.15);color:var(--critical)}.badge-high{background:rgba(255,136,0,0.15);color:var(--high)}.badge-medium{background:rgba(255,204,0,0.15);color:var(--medium)}.badge-low{background:rgba(0,204,255,0.15);color:var(--low)}.badge-info{background:rgba(136,136,136,0.15);color:var(--info)}
.finding-body{display:grid;grid-template-columns:100px 1fr;gap:8px 16px;font-size:14px}
.finding-body dt{color:var(--text-dim);font-weight:500}
.finding-body dd{word-break:break-word}
.finding-body dd a{color:var(--accent);text-decoration:none}
.extractor{background:var(--surface-2);border-radius:6px;padding:12px;margin-top:12px}
.extractor h4{font-size:12px;text-transform:uppercase;color:var(--text-dim);margin-bottom:8px}
.extractor code{font-family:'JetBrains Mono',monospace;font-size:12px;color:var(--accent)}
.remediations{background:var(--surface);border:1px solid var(--border);border-radius:10px;padding:24px}
.remediations h2{font-size:18px;margin-bottom:16px}
.remediations ol{padding-left:20px}
.remediations li{margin-bottom:10px}
footer{text-align:center;color:var(--text-dim);font-size:12px;padding:32px 0;border-top:1px solid var(--border);margin-top:32px}
@media(max-width:640px){.finding-body{grid-template-columns:1fr}.finding-header{flex-direction:column}}
</style>
</head>
<body>
<div class="container">
`)

	// Header
	sb.WriteString(`<header>
<div class="brand"><div class="brand-icon">CS</div><span style="color:var(--accent);font-weight:600">ClawSec</span></div>
<h1>Penetration Test Report</h1>
<p class="meta">Target: ` + html.EscapeString(target) + ` &bull; Generated: ` + time.Now().Format("2006-01-02 15:04:05") + `</p>
</header>
`)

	// Stats cards
	sb.WriteString(`<div class="summary-grid">
<div class="card"><h2>Total Findings</h2><div class="stat-number">` + fmt.Sprintf("%d", len(findings)) + `</div></div>
<div class="card"><h2>Risk Score</h2><div class="stat-number" style="color:` + calculateRiskColor(len(findings), severityCounts) + `">` + calculateRiskScore(len(findings), severityCounts) + `</div></div>
</div>
`)

	// Severity breakdown
	sb.WriteString(`<div class="card"><h2>Severity Breakdown</h2><div class="sev-grid">`)
	for _, sev := range severityOrder {
		count := severityCounts[sev]
		sb.WriteString(fmt.Sprintf(`<div class="sev-item %s"><div class="sev-count" style="color:%s">%d</div><div class="sev-label">%s</div></div>`,
			sev, severityColors[sev], count, strings.ToUpper(sev)))
	}
	sb.WriteString(`</div></div>`)

	// Findings
	sb.WriteString(`<section class="findings" style="margin-top:24px"><h2>Technical Findings (` + fmt.Sprintf("%d", len(findings)) + `)</h2>`)
	for i, f := range findings {
		sb.WriteString(`<div class="finding ` + f.Severity + `">
<div class="finding-header">
<span class="finding-title">` + fmt.Sprintf("#%d ", i+1) + html.EscapeString(f.Name) + `</span>
<span class="badge badge-` + f.Severity + `">` + strings.ToUpper(f.Severity) + `</span>
</div>
<dl class="finding-body">
<dt>Type</dt><dd>` + html.EscapeString(f.Type) + `</dd>
<dt>Host</dt><dd>` + html.EscapeString(f.Host) + `</dd>
`)
		if f.URL != "" {
			sb.WriteString(`<dt>URL</dt><dd><a href="` + html.EscapeString(f.URL) + `" target="_blank">` + html.EscapeString(f.URL) + `</a></dd>
`)
		}
		if f.Port > 0 {
			sb.WriteString(`<dt>Port</dt><dd>` + fmt.Sprintf("%d", f.Port) + `</dd>
`)
		}
		sb.WriteString(`<dt>Description</dt><dd>` + html.EscapeString(f.Message) + `</dd>
</dl>
`)
		if len(f.Extractor) > 0 {
			sb.WriteString(`<div class="extractor"><h4>Extracted Data</h4>`)
			for k, v := range f.Extractor {
				sb.WriteString(`<div><code>` + html.EscapeString(k) + `</code> = ` + html.EscapeString(fmt.Sprintf("%v", v)) + `</div>`)
			}
			sb.WriteString(`</div>`)
		}
		sb.WriteString(`</div>`)
	}
	sb.WriteString(`</section>`)

	// Remediations
	if len(findings) > 0 {
		sb.WriteString(`<section class="remediations" style="margin-top:24px"><h2>Remediation Recommendations</h2><ol>`)
		for _, f := range findings {
			rem := getRemediation(f.Severity, f.Type)
			sb.WriteString(`<li><strong>` + html.EscapeString(f.Name) + `:</strong> ` + html.EscapeString(rem) + `</li>`)
		}
		sb.WriteString(`</ol></section>`)
	}

	// Footer
	sb.WriteString(`<footer>Generated by ClawSec &mdash; AI-Native Offensive Security Platform</footer>
</div></body></html>`)

	return sb.String()
}

func calculateRiskScore(total int, counts map[string]int) string {
	if total == 0 {
		return "A+"
	}
	score := counts["critical"]*10 + counts["high"]*5 + counts["medium"]*2 + counts["low"]
	if score >= 50 {
		return "F"
	} else if score >= 30 {
		return "D"
	} else if score >= 15 {
		return "C"
	} else if score >= 5 {
		return "B"
	}
	return "A"
}

func calculateRiskColor(total int, counts map[string]int) string {
	score := counts["critical"]*10 + counts["high"]*5 + counts["medium"]*2 + counts["low"]
	if score >= 30 {
		return "#ff4444"
	} else if score >= 15 {
		return "#ff8800"
	} else if score >= 5 {
		return "#ffcc00"
	}
	return "#22c55e"
}

func (r *ReportGenerator) generateJSON(findings []Result, target string) string {
	report := map[string]interface{}{
		"target":    target,
		"date":      time.Now().Format(time.RFC3339),
		"findings":  findings,
		"summary":   generateSummary(findings),
	}
	// Simple JSON serialization would be done by caller
	return fmt.Sprintf("%v", report)
}

func getSeverityEmoji(severity string) string {
	switch severity {
	case "critical":
		return "🔴"
	case "high":
		return "🟠"
	case "medium":
		return "🟡"
	case "low":
		return "🔵"
	default:
		return "⚪"
	}
}

func getRemediation(severity, vulnType string) string {
	remediations := map[string]string{
		"port":      "Review necessity of exposed service. Implement firewall rules or VPN access.",
		"vulnerability": "Apply vendor patch immediately. Implement virtual patch via WAF if immediate patching not possible.",
		"weak_password": "Enforce strong password policy. Implement MFA. Rotate compromised credentials.",
		"info":      "Review information disclosure. Minimize exposed metadata and version information.",
	}

	if r, ok := remediations[vulnType]; ok {
		return r
	}
	return "Review and remediate based on vendor recommendations and security best practices."
}

func generateSummary(findings []Result) map[string]interface{} {
	severityCounts := make(map[string]int)
	for _, f := range findings {
		severityCounts[f.Severity]++
	}
	return map[string]interface{}{
		"total":     len(findings),
		"severity":  severityCounts,
	}
}
