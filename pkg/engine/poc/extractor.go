package poc

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ExtractData extracts data from response using extractors
func ExtractData(extractors []Extractor, resp *ResponseData, vars map[string]interface{}) map[string]interface{} {
	results := make(map[string]interface{})

	for _, ext := range extractors {
		data := getPartData(ext.Part, resp)
		var extracted []string

		switch ext.Type {
		case "regex":
			extracted = extractRegex(ext, data)
		case "json":
			extracted = extractJSON(ext, data)
		case "kval":
			extracted = extractKVal(ext, resp)
		case "xpath":
			extracted = extractXPath(ext, data)
		case "dsl":
			extracted = extractDSL(ext, resp, vars)
		}

		if len(extracted) > 0 {
			name := ext.Name
			if name == "" {
				name = fmt.Sprintf("extract_%d", len(results))
			}
			if len(extracted) == 1 {
				results[name] = extracted[0]
			} else {
				results[name] = extracted
			}
		}
	}

	return results
}

func extractRegex(ext Extractor, data string) []string {
	var results []string
	for _, pattern := range ext.Regex {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		matches := re.FindAllStringSubmatch(data, -1)
		for _, match := range matches {
			group := ext.Group
			if group >= len(match) {
				group = 0
			}
			if match[group] != "" {
				results = append(results, match[group])
			}
		}
	}
	return results
}

func extractJSON(ext Extractor, data string) []string {
	var results []string
	var jsonData interface{}
	if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
		return nil
	}

	for _, path := range ext.JSON {
		val := getJSONPath(jsonData, path)
		if val != nil {
			results = append(results, fmt.Sprintf("%v", val))
		}
	}
	return results
}

func getJSONPath(data interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		if part == "" {
			continue
		}

		switch v := current.(type) {
		case map[string]interface{}:
			if val, ok := v[part]; ok {
				current = val
			} else {
				return nil
			}
		case []interface{}:
			// Try array index
			var idx int
			if _, err := fmt.Sscanf(part, "%d", &idx); err == nil && idx >= 0 && idx < len(v) {
				current = v[idx]
			} else {
				return nil
			}
		default:
			return nil
		}
	}

	return current
}

func extractKVal(ext Extractor, resp *ResponseData) []string {
	var results []string
	for _, key := range ext.KVal {
		if val := resp.GetHeader(key); val != "" {
			results = append(results, val)
		}
	}
	return results
}

func extractXPath(ext Extractor, data string) []string {
	// Simplified XPath - just extract content between tags
	var results []string
	for _, xpath := range ext.XPath {
		// Very basic tag extraction: //tagname -> find <tagname>...</tagname>
		tag := strings.TrimPrefix(xpath, "//")
		re := regexp.MustCompile(fmt.Sprintf(`<%s[^>]*>([^<]*)</%s>`, regexp.QuoteMeta(tag), regexp.QuoteMeta(tag)))
		matches := re.FindAllStringSubmatch(data, -1)
		for _, match := range matches {
			if len(match) > 1 {
				results = append(results, match[1])
			}
		}
	}
	return results
}

func extractDSL(ext Extractor, resp *ResponseData, vars map[string]interface{}) []string {
	var results []string
	for _, expr := range ext.DSL {
		result := EvaluateDSL(expr, resp, vars)
		if result != nil {
			results = append(results, fmt.Sprintf("%v", result))
		}
	}
	return results
}
