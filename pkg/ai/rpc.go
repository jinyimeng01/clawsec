package ai

// Request represents a JSON-RPC 2.0 request
type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      string      `json:"id,omitempty"`
}

// Response represents a JSON-RPC 2.0 response
type Response struct {
	JSONRPC string       `json:"jsonrpc"`
	Result  interface{}  `json:"result,omitempty"`
	Error   *RPCError    `json:"error,omitempty"`
	ID      string       `json:"id,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// AnalyzeResult is returned by analyze method
type AnalyzeResult struct {
	RiskScore            float64  `json:"risk_score"`
	AttackSurface        []string `json:"attack_surface"`
	RecommendedNextSteps []string `json:"recommended_next_steps"`
	HighValueTargets     []string `json:"high_value_targets"`
	Reasoning            string   `json:"reasoning"`
}

// SuggestResult is returned by suggest method
type SuggestResult struct {
	Templates     []TemplateSuggestion `json:"templates"`
	PriorityOrder []string             `json:"priority_order"`
}

// TemplateSuggestion represents a suggested PoC template
type TemplateSuggestion struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Severity   string  `json:"severity"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

// ChainResult is returned by chain method
type ChainResult struct {
	Steps             []ChainStep `json:"steps"`
	OverallConfidence float64     `json:"overall_confidence"`
}

// ChainStep represents one step in an exploit chain
type ChainStep struct {
	Step            int    `json:"step"`
	Action          string `json:"action"`
	Tool            string `json:"tool"`
	ExpectedOutcome string `json:"expected_outcome"`
	Fallback        string `json:"fallback"`
}

// ReportResult is returned by report method
type ReportResult struct {
	Report string `json:"report"`
	Format string `json:"format"`
}

// ChatResult is returned by chat method
type ChatResult struct {
	Response string `json:"response"`
}
