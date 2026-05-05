package poc

import (
	"fmt"
	"regexp"
	"strings"
)

// MatchResponse checks if response matches the given matchers
func MatchResponse(matchers []Matcher, resp *ResponseData, vars map[string]interface{}) (bool, map[string]bool) {
	if len(matchers) == 0 {
		return true, nil
	}

	results := make(map[string]bool)
	condition := "or"
	if len(matchers) > 0 && matchers[0].Condition != "" {
		condition = matchers[0].Condition
	}

	var matched bool
	for _, matcher := range matchers {
		m := matchSingle(matcher, resp, vars)
		if matcher.Name != "" {
			results[matcher.Name] = m
		}

		switch condition {
		case "and":
			if !m {
				return false, results
			}
			matched = true
		case "or":
			if m {
				return true, results
			}
		}
	}

	return matched, results
}

// MatchResponseWithCondition checks matchers with explicit condition
func MatchResponseWithCondition(matchers []Matcher, condition string, resp *ResponseData, vars map[string]interface{}) (bool, map[string]bool) {
	if len(matchers) == 0 {
		return true, nil
	}

	results := make(map[string]bool)
	var matched bool

	for _, matcher := range matchers {
		m := matchSingle(matcher, resp, vars)
		if matcher.Name != "" {
			results[matcher.Name] = m
		}

		switch condition {
		case "and":
			if !m {
				return false, results
			}
			matched = true
		default: // or
			if m {
				return true, results
			}
		}
	}

	return matched, results
}

func matchSingle(m Matcher, resp *ResponseData, vars map[string]interface{}) bool {
	data := getPartData(m.Part, resp)
	var matched bool

	switch m.Type {
	case "word":
		matched = matchWords(m, data)
	case "regex":
		matched = matchRegex(m, data)
	case "status":
		matched = matchStatus(m, resp)
	case "binary":
		matched = matchBinary(m, data)
	case "dsl":
		matched = matchDSL(m, resp, vars)
	case "size":
		matched = matchSize(m, len(data))
	default:
		matched = false
	}

	if m.Negative {
		matched = !matched
	}

	return matched
}

func matchWords(m Matcher, data string) bool {
	if len(m.Words) == 0 {
		return false
	}

	condition := m.Condition
	if condition == "" {
		condition = "or"
	}

	for _, word := range m.Words {
		found := strings.Contains(data, word)
		if condition == "or" && found {
			return true
		}
		if condition == "and" && !found {
			return false
		}
	}

	return condition == "and"
}

func matchRegex(m Matcher, data string) bool {
	if len(m.Regex) == 0 {
		return false
	}

	condition := m.Condition
	if condition == "" {
		condition = "or"
	}

	for _, pattern := range m.Regex {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		found := re.MatchString(data)
		if condition == "or" && found {
			return true
		}
		if condition == "and" && !found {
			return false
		}
	}

	return condition == "and"
}

func matchStatus(m Matcher, resp *ResponseData) bool {
	if len(m.Status) == 0 {
		return false
	}
	for _, status := range m.Status {
		if resp.StatusCode == status {
			return true
		}
	}
	return false
}

func matchBinary(m Matcher, data string) bool {
	// Simplified binary matching - compare hex strings
	for _, bin := range m.Binary {
		decoded, err := hexDecode(bin)
		if err != nil {
			continue
		}
		if strings.Contains(data, decoded) {
			return true
		}
	}
	return false
}

func matchDSL(m Matcher, resp *ResponseData, vars map[string]interface{}) bool {
	if len(m.DSL) == 0 {
		return false
	}

	condition := m.Condition
	if condition == "" {
		condition = "or"
	}

	for _, expr := range m.DSL {
		result := EvaluateDSL(expr, resp, vars)
		found := false
		if b, ok := result.(bool); ok {
			found = b
		}
		if condition == "or" && found {
			return true
		}
		if condition == "and" && !found {
			return false
		}
	}

	return condition == "and"
}

func matchSize(m Matcher, size int) bool {
	// Parse size expression like "100,200" or ">100"
	for _, s := range m.DSL {
		if s == fmt.Sprintf("%d", size) {
			return true
		}
	}
	return false
}

func getPartData(part string, resp *ResponseData) string {
	switch part {
	case "body", "":
		return resp.Body
	case "header":
		return resp.Headers
	case "status":
		return fmt.Sprintf("%d", resp.StatusCode)
	case "all":
		return resp.Headers + "\r\n\r\n" + resp.Body
	case "raw":
		return resp.Raw
	default:
		// Try to get specific header
		if strings.HasPrefix(part, "header_") {
			headerName := strings.TrimPrefix(part, "header_")
			return resp.GetHeader(headerName)
		}
		return resp.Body
	}
}

// ResponseData holds HTTP response information
type ResponseData struct {
	StatusCode int
	Headers    string
	Body       string
	Raw        string
	HeaderMap  map[string]string
	Duration   int64 // milliseconds
	Size       int
}

// GetHeader gets a specific header value
func (r *ResponseData) GetHeader(name string) string {
	if r.HeaderMap == nil {
		return ""
	}
	// Case-insensitive lookup
	for k, v := range r.HeaderMap {
		if strings.EqualFold(k, name) {
			return v
		}
	}
	return ""
}

func hexDecode(s string) (string, error) {
	if len(s)%2 != 0 {
		return "", fmt.Errorf("invalid hex string")
	}
	result := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		b, err := parseHexByte(s[i : i+2])
		if err != nil {
			return "", err
		}
		result[i/2] = b
	}
	return string(result), nil
}

func parseHexByte(s string) (byte, error) {
	var b byte
	for _, c := range s {
		var v byte
		switch {
		case c >= '0' && c <= '9':
			v = byte(c - '0')
		case c >= 'a' && c <= 'f':
			v = byte(c - 'a' + 10)
		case c >= 'A' && c <= 'F':
			v = byte(c - 'A' + 10)
		default:
			return 0, fmt.Errorf("invalid hex character")
		}
		b = b<<4 | v
	}
	return b, nil
}
