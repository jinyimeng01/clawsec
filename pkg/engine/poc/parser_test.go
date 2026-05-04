package poc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTemplate(t *testing.T) {
	tmpl, err := ParseTemplate("../../../templates/poc/test-apache-detect.yaml")
	if err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	if tmpl.ID != "test-apache-detect" {
		t.Errorf("ID = %q, want test-apache-detect", tmpl.ID)
	}
	if tmpl.Info.Name != "Apache HTTP Server Detection" {
		t.Errorf("Name = %q", tmpl.Info.Name)
	}
	if tmpl.Info.Severity != "info" {
		t.Errorf("Severity = %q, want info", tmpl.Info.Severity)
	}
	if len(tmpl.HTTP) != 1 {
		t.Fatalf("HTTP requests = %d, want 1", len(tmpl.HTTP))
	}
	if len(tmpl.HTTP[0].Matchers) != 1 {
		t.Errorf("Matchers = %d, want 1", len(tmpl.HTTP[0].Matchers))
	}
}

func TestParseTemplate_CVE(t *testing.T) {
	tmpl, err := ParseTemplate("../../../templates/poc/cve-2021-41773.yaml")
	if err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	if tmpl.ID != "CVE-2021-41773" {
		t.Errorf("ID = %q", tmpl.ID)
	}
	if tmpl.Info.Severity != "critical" {
		t.Errorf("Severity = %q, want critical", tmpl.Info.Severity)
	}
	if len(tmpl.Variables) == 0 {
		t.Errorf("Expected variables, got none")
	}
}

func TestLoadTemplatesFromDir(t *testing.T) {
	templates, err := LoadTemplatesFromDir("../../../templates/poc")
	if err != nil {
		t.Fatalf("LoadTemplatesFromDir failed: %v", err)
	}

	if len(templates) == 0 {
		t.Fatalf("Expected templates, got 0")
	}

	t.Logf("Loaded %d templates", len(templates))
	for _, tmpl := range templates {
		t.Logf("  - %s (%s)", tmpl.ID, tmpl.Info.Name)
	}
}

func TestFilterTemplates(t *testing.T) {
	templates := []*Template{
		{ID: "1", Info: TemplateInfo{Severity: "critical", Tags: []string{"cve", "rce"}}},
		{ID: "2", Info: TemplateInfo{Severity: "high", Tags: []string{"cve"}}},
		{ID: "3", Info: TemplateInfo{Severity: "medium", Tags: []string{"xss"}}},
		{ID: "4", Info: TemplateInfo{Severity: "info", Tags: []string{"tech"}}},
	}

	// Filter by severity
	filtered := FilterTemplates(templates, []string{"critical", "high"}, nil)
	if len(filtered) != 2 {
		t.Errorf("Severity filter: got %d, want 2", len(filtered))
	}

	// Filter by tags
	filtered = FilterTemplates(templates, nil, []string{"cve"})
	if len(filtered) != 2 {
		t.Errorf("Tag filter: got %d, want 2", len(filtered))
	}

	// Filter by both
	filtered = FilterTemplates(templates, []string{"critical"}, []string{"rce"})
	if len(filtered) != 1 {
		t.Errorf("Combined filter: got %d, want 1", len(filtered))
	}
}

func TestLoadTemplatesFromPaths(t *testing.T) {
	// Create temp dir with a test template
	tmpDir := t.TempDir()
	testYaml := `id: test-template
info:
  name: Test Template
  author: test
  severity: info
http:
  - method: GET
    path:
      - "{{BaseURL}}/"
    matchers:
      - type: status
        status:
          - 200
`
	testFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(testFile, []byte(testYaml), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	templates, err := LoadTemplatesFromPaths([]string{tmpDir})
	if err != nil {
		t.Fatalf("LoadTemplatesFromPaths failed: %v", err)
	}

	if len(templates) != 1 {
		t.Fatalf("Expected 1 template, got %d", len(templates))
	}

	if templates[0].ID != "test-template" {
		t.Errorf("ID = %q", templates[0].ID)
	}
}
