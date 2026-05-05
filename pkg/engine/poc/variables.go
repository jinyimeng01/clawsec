package poc

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html/template"
	"math/big"
	"net/url"
	"regexp"
	"strings"
)

var (
	// regex to match template variables like {{var_name}}
	templateVarRegex = regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`)
	// regex to match DSL functions like {{function(args)}}
	dslFuncRegex = regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\((.*?)\)\s*\}\}`)
)

// ExpandVariables expands template variables in a string
func ExpandVariables(input string, vars map[string]interface{}) string {
	if vars == nil {
		vars = make(map[string]interface{})
	}

	// First pass: expand simple variables
	result := templateVarRegex.ReplaceAllStringFunc(input, func(match string) string {
		// Extract variable name
		matches := templateVarRegex.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}
		name := matches[1]

		if val, ok := vars[name]; ok {
			return fmt.Sprintf("%v", val)
		}
		return match
	})

	// Second pass: expand DSL functions
	result = dslFuncRegex.ReplaceAllStringFunc(result, func(match string) string {
		matches := dslFuncRegex.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}
		funcName := matches[1]
		args := matches[2]
		return executeDSLFunction(funcName, args, vars)
	})

	return result
}

// InitTemplateVariables creates initial variables for a template
func InitTemplateVariables(baseURL string) map[string]interface{} {
	vars := make(map[string]interface{})
	vars["BaseURL"] = baseURL
	vars["Hostname"] = extractHostname(baseURL)
	vars["Host"] = extractHost(baseURL)
	vars["Port"] = extractPort(baseURL)
	vars["Path"] = extractPath(baseURL)
	vars["Scheme"] = extractScheme(baseURL)

	// Random generators
	vars["randstr"] = randomString(8)
	vars["randbase"] = randomString(8)
	vars["rand_text_alpha"] = randomAlphaString(8)
	vars["rand_text_numeric"] = randomNumericString(8)
	vars["rand_text_alphanumeric"] = randomAlphaNumericString(8)
	vars["rand_int"] = randomInt(1000, 9999)
	vars["rand_ip"] = randomIP()

	return vars
}

// MergeVariables merges multiple variable maps (later maps override earlier)
func MergeVariables(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// executeDSLFunction executes a simple DSL function
func executeDSLFunction(name, args string, vars map[string]interface{}) string {
	switch name {
	case "base64", "b64encode":
		return base64.StdEncoding.EncodeToString([]byte(args))
	case "base64_decode", "b64decode":
		data, _ := base64.StdEncoding.DecodeString(args)
		return string(data)
	case "md5":
		hash := md5.Sum([]byte(args))
		return hex.EncodeToString(hash[:])
	case "sha1":
		hash := sha1.Sum([]byte(args))
		return hex.EncodeToString(hash[:])
	case "sha256":
		hash := sha256.Sum256([]byte(args))
		return hex.EncodeToString(hash[:])
	case "hex_encode", "hexencode":
		return hex.EncodeToString([]byte(args))
	case "hex_decode", "hexdecode":
		data, _ := hex.DecodeString(args)
		return string(data)
	case "url_encode", "urlencode":
		return url.QueryEscape(args)
	case "url_decode", "urldecode":
		result, _ := url.QueryUnescape(args)
		return result
	case "to_lower", "lower":
		return strings.ToLower(args)
	case "to_upper", "upper":
		return strings.ToUpper(args)
	case "trim", "trim_space":
		return strings.TrimSpace(args)
	case "len", "length":
		return fmt.Sprintf("%d", len(args))
	case "rand_base":
		length := 8
		fmt.Sscanf(args, "%d", &length)
		return randomString(length)
	case "rand_text":
		length := 8
		fmt.Sscanf(args, "%d", &length)
		return randomAlphaNumericString(length)
	case "rand_int":
		min, max := 1, 1000
		fmt.Sscanf(args, "%d,%d", &min, &max)
		return fmt.Sprintf("%d", randomInt(min, max))
	case "rand_ip":
		return randomIP()
	case "replace":
		parts := strings.SplitN(args, ",", 3)
		if len(parts) == 3 {
			return strings.ReplaceAll(parts[0], parts[1], parts[2])
		}
	case "contains":
		parts := strings.SplitN(args, ",", 2)
		if len(parts) == 2 {
			return fmt.Sprintf("%v", strings.Contains(parts[0], parts[1]))
		}
	case "concat":
		return strings.ReplaceAll(args, ",", "")
	case "join":
		parts := strings.SplitN(args, ",", 2)
		if len(parts) == 2 {
			sep := parts[0]
			items := strings.Split(parts[1], ",")
			return strings.Join(items, sep)
		}
	case "reverse":
		runes := []rune(args)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	case "html_escape":
		return template.HTMLEscapeString(args)
	case "html_unescape":
		return string(template.HTML(args))
	}

	// Try to resolve as variable
	if val, ok := vars[name]; ok {
		return fmt.Sprintf("%v", val)
	}

	return fmt.Sprintf("{{%s(%s)}}", name, args)
}

// Helper functions
func extractHostname(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return u.Hostname()
}

func extractHost(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return u.Host
}

func extractPort(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	if u.Port() != "" {
		return u.Port()
	}
	if u.Scheme == "https" {
		return "443"
	}
	return "80"
}

func extractPath(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "/"
	}
	if u.Path == "" {
		return "/"
	}
	return u.Path
}

func extractScheme(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "http"
	}
	if u.Scheme == "" {
		return "http"
	}
	return u.Scheme
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[n.Int64()]
	}
	return string(result)
}

func randomAlphaString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[n.Int64()]
	}
	return string(result)
}

func randomNumericString(length int) string {
	const charset = "0123456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[n.Int64()]
	}
	return string(result)
}

func randomAlphaNumericString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[n.Int64()]
	}
	return string(result)
}

func randomInt(min, max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	return int(n.Int64()) + min
}

func randomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		randomInt(1, 254),
		randomInt(0, 255),
		randomInt(0, 255),
		randomInt(1, 254),
	)
}
