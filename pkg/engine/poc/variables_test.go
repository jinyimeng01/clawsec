package poc

import (
	"testing"
)

func TestExpandVariables(t *testing.T) {
	vars := map[string]interface{}{
		"BaseURL":  "http://example.com",
		"Hostname": "example.com",
		"test_var": "hello",
	}

	tests := []struct {
		input string
		want  string
	}{
		{"{{BaseURL}}/api", "http://example.com/api"},
		{"{{Hostname}}:8080", "example.com:8080"},
		{"{{test_var}} world", "hello world"},
		{"/path/{{test_var}}/test", "/path/hello/test"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ExpandVariables(tt.input, vars)
			if got != tt.want {
				t.Errorf("ExpandVariables(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestInitTemplateVariables(t *testing.T) {
	vars := InitTemplateVariables("https://example.com:8443/path")

	if vars["BaseURL"] != "https://example.com:8443/path" {
		t.Errorf("BaseURL = %v, want %v", vars["BaseURL"], "https://example.com:8443/path")
	}
	if vars["Hostname"] != "example.com" {
		t.Errorf("Hostname = %v", vars["Hostname"])
	}
	if vars["Port"] != "8443" {
		t.Errorf("Port = %v, want 8443", vars["Port"])
	}
	if vars["Scheme"] != "https" {
		t.Errorf("Scheme = %v, want https", vars["Scheme"])
	}
	if vars["Path"] != "/path" {
		t.Errorf("Path = %v, want /path", vars["Path"])
	}
}

func TestExecuteDSLFunction(t *testing.T) {
	vars := map[string]interface{}{}

	tests := []struct {
		name string
		args string
		want string
	}{
		{"base64", "hello", "aGVsbG8="},
		{"base64_decode", "aGVsbG8=", "hello"},
		{"to_lower", "HELLO", "hello"},
		{"to_upper", "hello", "HELLO"},
		{"md5", "hello", "5d41402abc4b2a76b9719d911017c592"},
		{"hex_encode", "hello", "68656c6c6f"},
		{"url_encode", "hello world", "hello+world"},
		{"trim", "  hello  ", "hello"},
		{"replace", "hello,el,xx", "hxxlo"},
		{"concat", "a,b,c", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := executeDSLFunction(tt.name, tt.args, vars)
			if got != tt.want {
				t.Errorf("executeDSLFunction(%q, %q) = %q, want %q", tt.name, tt.args, got, tt.want)
			}
		})
	}
}

func TestMergeVariables(t *testing.T) {
	m1 := map[string]interface{}{"a": "1", "b": "2"}
	m2 := map[string]interface{}{"b": "3", "c": "4"}

	result := MergeVariables(m1, m2)
	if result["a"] != "1" || result["b"] != "3" || result["c"] != "4" {
		t.Errorf("MergeVariables failed: %v", result)
	}
}

func BenchmarkExpandVariables(b *testing.B) {
	vars := InitTemplateVariables("https://example.com:8443/api")
	input := "{{BaseURL}}/{{randstr}}/test?host={{Hostname}}"

	for i := 0; i < b.N; i++ {
		_ = ExpandVariables(input, vars)
	}
}
