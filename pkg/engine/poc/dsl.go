package poc

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// EvaluateDSL evaluates a DSL expression
func EvaluateDSL(expr string, resp *ResponseData, vars map[string]interface{}) interface{} {
	expr = strings.TrimSpace(expr)

	// Replace response variable references
	expr = replaceResponseVars(expr, resp)

	// Evaluate common DSL patterns
	if result := evalContains(expr); result != nil {
		return *result
	}
	if result := evalCompare(expr); result != nil {
		return *result
	}
	if result := evalStatusCode(expr, resp); result != nil {
		return *result
	}
	if result := evalRegex(expr, resp); result != nil {
		return *result
	}
	if result := evalSize(expr, resp); result != nil {
		return *result
	}
	if result := evalAndOr(expr, resp, vars); result != nil {
		return *result
	}

	// Fallback: try to evaluate as boolean string
	if expr == "true" {
		return true
	}
	if expr == "false" {
		return false
	}

	return nil
}

func replaceResponseVars(expr string, resp *ResponseData) string {
	expr = strings.ReplaceAll(expr, "status_code", fmt.Sprintf("%d", resp.StatusCode))
	expr = strings.ReplaceAll(expr, "status", fmt.Sprintf("%d", resp.StatusCode))
	expr = strings.ReplaceAll(expr, "body", fmt.Sprintf("%q", resp.Body))
	expr = strings.ReplaceAll(expr, "headers", fmt.Sprintf("%q", resp.Headers))
	expr = strings.ReplaceAll(expr, "raw", fmt.Sprintf("%q", resp.Raw))
	expr = strings.ReplaceAll(expr, "size", fmt.Sprintf("%d", resp.Size))
	expr = strings.ReplaceAll(expr, "duration", fmt.Sprintf("%d", resp.Duration))
	return expr
}

func evalContains(expr string) *bool {
	// Pattern: contains(body, "string") or contains(headers, "string")
	re := regexp.MustCompile(`contains\s*\(\s*([^,]+)\s*,\s*"([^"]*)"\s*\)`)
	matches := re.FindAllStringSubmatch(expr, -1)
	if len(matches) == 0 {
		return nil
	}

	result := true
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		data := strings.TrimSpace(match[1])
		search := match[2]
		if !strings.Contains(data, search) {
			result = false
			break
		}
	}
	return &result
}

func evalCompare(expr string) *bool {
	// Pattern: "string" == "string" or number == number
	re := regexp.MustCompile(`(?:(\d+)|"([^"]*)")\s*(==|!=|>=|<=|>|<)\s*(?:(\d+)|"([^"]*)")`)
	match := re.FindStringSubmatch(expr)
	if match == nil {
		return nil
	}

	leftNum := match[1]
	leftStr := match[2]
	op := match[3]
	rightNum := match[4]
	rightStr := match[5]

	if leftNum != "" && rightNum != "" {
		l, _ := strconv.Atoi(leftNum)
		r, _ := strconv.Atoi(rightNum)
		result := compareInt(l, r, op)
		return &result
	}

	result := compareStr(leftStr, rightStr, op)
	return &result
}

func evalStatusCode(expr string, resp *ResponseData) *bool {
	re := regexp.MustCompile(`status_code\s*(==|!=)\s*(\d+)`)
	match := re.FindStringSubmatch(expr)
	if match == nil {
		return nil
	}

	op := match[1]
	code, _ := strconv.Atoi(match[2])
	result := false
	if op == "==" {
		result = resp.StatusCode == code
	} else if op == "!=" {
		result = resp.StatusCode != code
	}
	return &result
}

func evalRegex(expr string, resp *ResponseData) *bool {
	re := regexp.MustCompile(`regex\s*\(\s*([^,]+)\s*,\s*"([^"]*)"\s*\)`)
	matches := re.FindAllStringSubmatch(expr, -1)
	if len(matches) == 0 {
		return nil
	}

	result := true
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		data := strings.TrimSpace(match[1])
		pattern := match[2]
		re2, err := regexp.Compile(pattern)
		if err != nil {
			result = false
			break
		}
		if !re2.MatchString(data) {
			result = false
			break
		}
	}
	return &result
}

func evalSize(expr string, resp *ResponseData) *bool {
	re := regexp.MustCompile(`size\s*(==|>=|<=|>|<)\s*(\d+)`)
	match := re.FindStringSubmatch(expr)
	if match == nil {
		return nil
	}

	op := match[1]
	size, _ := strconv.Atoi(match[2])
	result := compareInt(resp.Size, size, op)
	return &result
}

func evalAndOr(expr string, resp *ResponseData, vars map[string]interface{}) *bool {
	// Handle && and ||
	if strings.Contains(expr, "&&") {
		parts := strings.Split(expr, "&&")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			result := EvaluateDSL(part, resp, vars)
			if b, ok := result.(bool); !ok || !b {
				f := false
				return &f
			}
		}
		t := true
		return &t
	}

	if strings.Contains(expr, "||") {
		parts := strings.Split(expr, "||")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			result := EvaluateDSL(part, resp, vars)
			if b, ok := result.(bool); ok && b {
				t := true
				return &t
			}
		}
		f := false
		return &f
	}

	return nil
}

func compareInt(a, b int, op string) bool {
	switch op {
	case "==":
		return a == b
	case "!=":
		return a != b
	case ">=":
		return a >= b
	case "<=":
		return a <= b
	case ">":
		return a > b
	case "<":
		return a < b
	default:
		return false
	}
}

func compareStr(a, b, op string) bool {
	switch op {
	case "==":
		return a == b
	case "!=":
		return a != b
	case ">=":
		return a >= b
	case "<=":
		return a <= b
	case ">":
		return a > b
	case "<":
		return a < b
	default:
		return false
	}
}
