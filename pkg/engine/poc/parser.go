package poc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseTemplate parses a single template file
func ParseTemplate(path string) (*Template, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template %s: %w", path, err)
	}

	var tmpl Template
	if err := yaml.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", path, err)
	}

	tmpl.Path = path

	// Validate
	if tmpl.ID == "" {
		return nil, fmt.Errorf("template %s: missing id", path)
	}
	if tmpl.Info.Name == "" {
		return nil, fmt.Errorf("template %s: missing info.name", path)
	}

	// Normalize severity
	tmpl.Info.Severity = strings.ToLower(tmpl.Info.Severity)

	// Set default matchers-condition
	for i := range tmpl.HTTP {
		if tmpl.HTTP[i].MatchersCondition == "" {
			tmpl.HTTP[i].MatchersCondition = "or"
		}
	}
	for i := range tmpl.TCP {
		if tmpl.TCP[i].MatchersCondition == "" {
			tmpl.TCP[i].MatchersCondition = "or"
		}
	}
	for i := range tmpl.DNS {
		if tmpl.DNS[i].MatchersCondition == "" {
			tmpl.DNS[i].MatchersCondition = "or"
		}
	}
	for i := range tmpl.SSL {
		if tmpl.SSL[i].MatchersCondition == "" {
			tmpl.SSL[i].MatchersCondition = "or"
		}
	}

	return &tmpl, nil
}

// LoadTemplatesFromDir loads all templates from a directory recursively
func LoadTemplatesFromDir(dir string) ([]*Template, error) {
	var templates []*Template

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on error
		}
		if info.IsDir() {
			return nil
		}
		if !isTemplateFile(path) {
			return nil
		}

		tmpl, err := ParseTemplate(path)
		if err != nil {
			// Continue - don't fail entire load for one bad template
			return nil
		}
		templates = append(templates, tmpl)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return templates, nil
}

// LoadTemplatesFromPaths loads templates from multiple file paths
func LoadTemplatesFromPaths(paths []string) ([]*Template, error) {
	var templates []*Template
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("stat failed for %s: %w", path, err)
		}

		if info.IsDir() {
			dirTemplates, err := LoadTemplatesFromDir(path)
			if err != nil {
				return nil, fmt.Errorf("load dir failed for %s: %w", path, err)
			}
			templates = append(templates, dirTemplates...)
		} else {
			tmpl, err := ParseTemplate(path)
			if err != nil {
				return nil, fmt.Errorf("parse failed for %s: %w", path, err)
			}
			templates = append(templates, tmpl)
		}
	}
	return templates, nil
}

// FilterTemplates filters templates by severity and tags
func FilterTemplates(templates []*Template, severities, tags []string) []*Template {
	if len(severities) == 0 && len(tags) == 0 {
		return templates
	}

	// Normalize filters
	severitySet := make(map[string]bool)
	for _, s := range severities {
		severitySet[strings.ToLower(s)] = true
	}
	tagSet := make(map[string]bool)
	for _, t := range tags {
		for _, part := range strings.Split(t, ",") {
			tagSet[strings.TrimSpace(strings.ToLower(part))] = true
		}
	}

	var filtered []*Template
	for _, tmpl := range templates {
		// Check severity
		if len(severitySet) > 0 {
			if !severitySet[tmpl.Info.Severity] {
				continue
			}
		}

		// Check tags
		if len(tagSet) > 0 {
			matched := false
			for _, tmplTag := range tmpl.Info.Tags {
				for _, part := range strings.Split(tmplTag, ",") {
					if tagSet[strings.TrimSpace(strings.ToLower(part))] {
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}
			if !matched {
				continue
			}
		}

		filtered = append(filtered, tmpl)
	}

	return filtered
}

func isTemplateFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}
