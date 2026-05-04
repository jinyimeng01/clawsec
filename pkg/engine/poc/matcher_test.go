package poc

import (
	"testing"
)

func TestMatchWords(t *testing.T) {
	resp := &ResponseData{
		StatusCode: 200,
		Body:       "Welcome to Apache server",
		Headers:    "Server: Apache/2.4.41",
	}

	tests := []struct {
		name    string
		matcher Matcher
		want    bool
	}{
		{
			name: "simple word match",
			matcher: Matcher{
				Type:  "word",
				Words: []string{"Apache"},
				Part:  "body",
			},
			want: true,
		},
		{
			name: "multiple words OR",
			matcher: Matcher{
				Type:      "word",
				Words:     []string{"nginx", "Apache"},
				Part:      "body",
				Condition: "or",
			},
			want: true,
		},
		{
			name: "multiple words AND",
			matcher: Matcher{
				Type:      "word",
				Words:     []string{"Welcome", "Apache"},
				Part:      "body",
				Condition: "and",
			},
			want: true,
		},
		{
			name: "header match",
			matcher: Matcher{
				Type:  "word",
				Words: []string{"Apache/2.4"},
				Part:  "header",
			},
			want: true,
		},
		{
			name: "negative match",
			matcher: Matcher{
				Type:     "word",
				Words:    []string{"nginx"},
				Part:     "body",
				Negative: true,
			},
			want: true,
		},
		{
			name: "no match",
			matcher: Matcher{
				Type:  "word",
				Words: []string{"nginx"},
				Part:  "body",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := MatchResponse([]Matcher{tt.matcher}, resp, nil)
			if got != tt.want {
				t.Errorf("MatchResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchStatus(t *testing.T) {
	resp := &ResponseData{StatusCode: 200}

	matcher := Matcher{
		Type:   "status",
		Status: []int{200, 301},
	}

	got, _ := MatchResponse([]Matcher{matcher}, resp, nil)
	if !got {
		t.Errorf("MatchResponse() = %v, want true", got)
	}

	resp.StatusCode = 404
	got, _ = MatchResponse([]Matcher{matcher}, resp, nil)
	if got {
		t.Errorf("MatchResponse() = %v, want false", got)
	}
}

func TestMatchRegex(t *testing.T) {
	resp := &ResponseData{Body: "Apache/2.4.41 (Ubuntu)"}

	matcher := Matcher{
		Type:  "regex",
		Regex: []string{`Apache/\d+\.\d+\.\d+`},
		Part:  "body",
	}

	got, _ := MatchResponse([]Matcher{matcher}, resp, nil)
	if !got {
		t.Errorf("MatchResponse() = %v, want true", got)
	}
}

func TestMatchDSL(t *testing.T) {
	resp := &ResponseData{
		StatusCode: 200,
		Body:       "Hello World",
		Size:       11,
	}

	tests := []struct {
		name    string
		expr    string
		want    bool
	}{
		{"status check", "status_code == 200", true},
		{"body contains", `contains("Hello World", "World")`, true},
		{"combined", `status_code == 200 && contains("Hello World", "Hello")`, true},
		{"size check", "size >= 10", true},
		{"false condition", "status_code == 404", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := Matcher{
				Type: "dsl",
				DSL:  []string{tt.expr},
			}
			got, _ := MatchResponse([]Matcher{matcher}, resp, nil)
			if got != tt.want {
				t.Errorf("MatchResponse(%q) = %v, want %v", tt.expr, got, tt.want)
			}
		})
	}
}

func TestMatchResponseWithCondition(t *testing.T) {
	resp := &ResponseData{
		StatusCode: 200,
		Body:       "Apache server",
	}

	matchers := []Matcher{
		{Type: "status", Status: []int{200}},
		{Type: "word", Words: []string{"Apache"}, Part: "body"},
	}

	// AND condition
	got, _ := MatchResponseWithCondition(matchers, "and", resp, nil)
	if !got {
		t.Errorf("AND condition = %v, want true", got)
	}

	// OR condition
	got, _ = MatchResponseWithCondition(matchers, "or", resp, nil)
	if !got {
		t.Errorf("OR condition = %v, want true", got)
	}
}
