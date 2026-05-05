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
	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Penetration Test Report</title>
<style>
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 900px; margin: 0 auto; padding: 20px; background: #f5f5f5; }
.header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; border-radius: 8px; margin-bottom: 20px; }
.finding { background: white; border-radius: 8px; padding: 20px; margin-bottom: 15px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
.severity-critical { border-left: 4px solid #dc3545; }
.severity-high { border-left: 4px solid #fd7e14; }
.severity-medium { border-left: 4px solid #ffc107; }
.severity-low { border-left: 4px solid #17a2b8; }
.severity-info { border-left: 4px solid #6c757d; }
.summary { background: white; border-radius: 8px; padding: 20px; margin-bottom: 20px; }
.badge { display: inline-block; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; margin-right: 8px; }
.badge-critical { background: #dc3545; color: white; }
.badge-high { background: #fd7e14; color: white; }
.badge-medium { background: #ffc107; color: black; }
.badge-low { background: #17a2b8; color: white; }
.badge-info { background: #6c757d; color: white; }
table { width: 100%; border-collapse: collapse; margin: 15px 0; }
th, td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
th { background: #f8f9fa; font-weight: 600; }
</style>
</head>
<body>
`)

	sb.WriteString(`<div class="header">
<h1>Penetration Test Report</h1>
<p>Target: ` + html.EscapeString(target) + `</p>
<p>Date: ` + time.Now().Format("2006-01-02") + `</p>
</div>
`)

	// Summary
	sb.WriteString(`<div class="summary">
<h2>Executive Summary</h2>
`)
	if len(findings) == 0 {
		sb.WriteString("<p>No vulnerabilities were identified during this assessment.</p>")
	} else {
		sb.WriteString(fmt.Sprintf("<p>A total of <strong>%d</strong> findings were identified:</p>", len(findings)))
		sb.WriteString("<table><tr><th>Severity</th><th>Count</th></tr>")
		severityCounts := make(map[string]int)
		for _, f := range findings {
			severityCounts[f.Severity]++
		}
		for sev, count := range severityCounts {
			sb.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%d</td></tr>", strings.ToUpper(sev), count))
		}
		sb.WriteString("</table>")
	}
	sb.WriteString("</div>")

	// Findings
	sb.WriteString("<h2>Technical Findings</h2>")
	for _, f := range findings {
		sevClass := "severity-" + f.Severity
		badgeClass := "badge-" + f.Severity
		sb.WriteString(fmt.Sprintf(`<div class="finding %s">
<h3>%s <span class="badge %s">%s</span></h3>
<p><strong>Type:</strong> %s</p>
<p><strong>Host:</strong> %s</p>
<p><strong>Description:</strong> %s</p>
</div>
`, sevClass, html.EscapeString(f.Name), badgeClass, strings.ToUpper(f.Severity),
			html.EscapeString(f.Type), html.EscapeString(f.Host), html.EscapeString(f.Message)))
	}

	sb.WriteString("</body></html>")
	return sb.String()
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
