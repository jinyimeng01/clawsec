package poc

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// StringOrSlice can unmarshal from either a string or a slice of strings
type StringOrSlice []string

// UnmarshalYAML implements custom YAML unmarshaling
func (s *StringOrSlice) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		// Single string value - split by comma
		*s = strings.Split(node.Value, ",")
		for i := range *s {
			(*s)[i] = strings.TrimSpace((*s)[i])
		}
		return nil
	case yaml.SequenceNode:
		// Array of strings
		var result []string
		for _, item := range node.Content {
			if item.Kind == yaml.ScalarNode {
				result = append(result, item.Value)
			}
		}
		*s = result
		return nil
	default:
		return fmt.Errorf("cannot unmarshal %v into StringOrSlice", node.Kind)
	}
}

// Template represents a Nuclei-compatible PoC template
type Template struct {
	ID        string            `yaml:"id" json:"id"`
	Info      TemplateInfo      `yaml:"info" json:"info"`
	Variables map[string]string `yaml:"variables,omitempty" json:"variables,omitempty"`
	HTTP      []HTTPRequest     `yaml:"http,omitempty" json:"http,omitempty"`
	TCP       []TCPRequest      `yaml:"tcp,omitempty" json:"tcp,omitempty"`
	DNS       []DNSRequest      `yaml:"dns,omitempty" json:"dns,omitempty"`
	SSL       []SSLRequest      `yaml:"ssl,omitempty" json:"ssl,omitempty"`
	Workflow  []WorkflowStep    `yaml:"workflow,omitempty" json:"workflow,omitempty"`
	Path      string            `yaml:"-" json:"-"`
}

// TemplateInfo holds template metadata
type TemplateInfo struct {
	Name           string            `yaml:"name" json:"name"`
	Author         string            `yaml:"author" json:"author"`
	Severity       string            `yaml:"severity" json:"severity"`
	Description    string            `yaml:"description,omitempty" json:"description,omitempty"`
	Reference      StringOrSlice     `yaml:"reference,omitempty" json:"reference,omitempty"`
	Tags           StringOrSlice     `yaml:"tags,omitempty" json:"tags,omitempty"`
	Metadata       map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Classification *Classification   `yaml:"classification,omitempty" json:"classification,omitempty"`
}

// Classification holds vulnerability classification
type Classification struct {
	CVSSScore   float64  `yaml:"cvss-score,omitempty" json:"cvss-score,omitempty"`
	CVSSMetrics string   `yaml:"cvss-metrics,omitempty" json:"cvss-metrics,omitempty"`
	CWEID       []string `yaml:"cwe-id,omitempty" json:"cwe-id,omitempty"`
	CVEID       []string `yaml:"cve-id,omitempty" json:"cve-id,omitempty"`
	EPSSScore   float64  `yaml:"epss-score,omitempty" json:"epss-score,omitempty"`
}

// HTTPRequest represents an HTTP request block
type HTTPRequest struct {
	Method            string            `yaml:"method,omitempty" json:"method,omitempty"`
	Path              []string          `yaml:"path" json:"path"`
	Headers           map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	Body              string            `yaml:"body,omitempty" json:"body,omitempty"`
	Raw               []string          `yaml:"raw,omitempty" json:"raw,omitempty"`
	MatchersCondition string            `yaml:"matchers-condition,omitempty" json:"matchers-condition,omitempty"`
	Matchers          []Matcher         `yaml:"matchers" json:"matchers"`
	Extractors        []Extractor       `yaml:"extractors,omitempty" json:"extractors,omitempty"`
	StopAtFirstMatch  bool              `yaml:"stop-at-first-match,omitempty" json:"stop-at-first-match,omitempty"`
	CookieReuse       bool              `yaml:"cookie-reuse,omitempty" json:"cookie-reuse,omitempty"`
	Redirects         bool              `yaml:"redirects,omitempty" json:"redirects,omitempty"`
	MaxRedirects      int               `yaml:"max-redirects,omitempty" json:"max-redirects,omitempty"`
}

// TCPRequest represents a TCP request block
type TCPRequest struct {
	Host              []string    `yaml:"host,omitempty" json:"host,omitempty"`
	Inputs            []TCPInput  `yaml:"inputs" json:"inputs"`
	MatchersCondition string      `yaml:"matchers-condition,omitempty" json:"matchers-condition,omitempty"`
	Matchers          []Matcher   `yaml:"matchers" json:"matchers"`
	Extractors        []Extractor `yaml:"extractors,omitempty" json:"extractors,omitempty"`
}

// TCPInput represents a TCP input
type TCPInput struct {
	Data string `yaml:"data" json:"data"`
	Type string `yaml:"type,omitempty" json:"type,omitempty"`
}

// DNSRequest represents a DNS request block
type DNSRequest struct {
	Name              []string    `yaml:"name" json:"name"`
	Type              string      `yaml:"type" json:"type"`
	Class             string      `yaml:"class,omitempty" json:"class,omitempty"`
	Retries           int         `yaml:"retries,omitempty" json:"retries,omitempty"`
	MatchersCondition string      `yaml:"matchers-condition,omitempty" json:"matchers-condition,omitempty"`
	Matchers          []Matcher   `yaml:"matchers" json:"matchers"`
	Extractors        []Extractor `yaml:"extractors,omitempty" json:"extractors,omitempty"`
}

// SSLRequest represents an SSL request block
type SSLRequest struct {
	Address           []string    `yaml:"address" json:"address"`
	MatchersCondition string      `yaml:"matchers-condition,omitempty" json:"matchers-condition,omitempty"`
	Matchers          []Matcher   `yaml:"matchers" json:"matchers"`
	Extractors        []Extractor `yaml:"extractors,omitempty" json:"extractors,omitempty"`
}

// Matcher represents a response matcher
type Matcher struct {
	Type      string   `yaml:"type" json:"type"`
	Part      string   `yaml:"part,omitempty" json:"part,omitempty"`
	Words     []string `yaml:"words,omitempty" json:"words,omitempty"`
	Regex     []string `yaml:"regex,omitempty" json:"regex,omitempty"`
	DSL       []string `yaml:"dsl,omitempty" json:"dsl,omitempty"`
	Status    []int    `yaml:"status,omitempty" json:"status,omitempty"`
	Binary    []string `yaml:"binary,omitempty" json:"binary,omitempty"`
	Condition string   `yaml:"condition,omitempty" json:"condition,omitempty"`
	Negative  bool     `yaml:"negative,omitempty" json:"negative,omitempty"`
	Name      string   `yaml:"name,omitempty" json:"name,omitempty"`
	Encoding  string   `yaml:"encoding,omitempty" json:"encoding,omitempty"`
}

// Extractor represents a data extractor
type Extractor struct {
	Type      string   `yaml:"type" json:"type"`
	Part      string   `yaml:"part,omitempty" json:"part,omitempty"`
	Name      string   `yaml:"name,omitempty" json:"name,omitempty"`
	Internal  bool     `yaml:"internal,omitempty" json:"internal,omitempty"`
	Group     int      `yaml:"group,omitempty" json:"group,omitempty"`
	Regex     []string `yaml:"regex,omitempty" json:"regex,omitempty"`
	JSON      []string `yaml:"json,omitempty" json:"json,omitempty"`
	XPath     []string `yaml:"xpath,omitempty" json:"xpath,omitempty"`
	Attribute string   `yaml:"attribute,omitempty" json:"attribute,omitempty"`
	DSL       []string `yaml:"dsl,omitempty" json:"dsl,omitempty"`
	KVal      []string `yaml:"kval,omitempty" json:"kval,omitempty"`
}

// WorkflowStep represents a workflow step
type WorkflowStep struct {
	Template     string            `yaml:"template" json:"template"`
	Matchers     []WorkflowMatcher `yaml:"matchers,omitempty" json:"matchers,omitempty"`
	Subtemplates []string          `yaml:"subtemplates,omitempty" json:"subtemplates,omitempty"`
	Variables    map[string]string `yaml:"variables,omitempty" json:"variables,omitempty"`
}

// WorkflowMatcher represents a workflow matcher
type WorkflowMatcher struct {
	Name string `yaml:"name" json:"name"`
}

// Result represents a template execution result
type Result struct {
	TemplateID       string                 `json:"template-id"`
	TemplatePath     string                 `json:"template-path"`
	Info             TemplateInfo           `json:"info"`
	Type             string                 `json:"type"`
	Host             string                 `json:"host"`
	Port             string                 `json:"port,omitempty"`
	URL              string                 `json:"url,omitempty"`
	MatchedAt        time.Time              `json:"matched-at"`
	MatcherName      string                 `json:"matcher-name,omitempty"`
	ExtractedResults []string               `json:"extracted-results,omitempty"`
	Request          string                 `json:"request,omitempty"`
	Response         string                 `json:"response,omitempty"`
	CurlCommand      string                 `json:"curl-command,omitempty"`
	Meta             map[string]interface{} `json:"meta,omitempty"`
}

// ExecutionContext holds runtime execution state
type ExecutionContext struct {
	BaseURL   string
	Host      string
	Port      int
	Scheme    string
	Variables map[string]interface{}
	Extracted map[string]interface{}
	Cookies   map[string]string
	Template  *Template
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(baseURL string) *ExecutionContext {
	return &ExecutionContext{
		BaseURL:   baseURL,
		Variables: make(map[string]interface{}),
		Extracted: make(map[string]interface{}),
		Cookies:   make(map[string]string),
	}
}

// SeverityScore returns numeric severity score for sorting
func SeverityScore(s string) int {
	switch s {
	case "critical":
		return 5
	case "high":
		return 4
	case "medium":
		return 3
	case "low":
		return 2
	case "info":
		return 1
	default:
		return 0
	}
}
