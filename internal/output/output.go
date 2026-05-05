package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// Format represents output format
type Format string

const (
	FormatText     Format = "text"
	FormatJSON     Format = "json"
	FormatJSONL    Format = "jsonl"
	FormatCSV      Format = "csv"
	FormatMarkdown Format = "markdown"
	FormatHTML     Format = "html"
	FormatSilent   Format = "silent"
)

// Result represents a generic scan result
type Result struct {
	Timestamp   time.Time              `json:"timestamp"`
	Type        string                 `json:"type"`
	Level       string                 `json:"level"`
	Host        string                 `json:"host,omitempty"`
	Port        int                    `json:"port,omitempty"`
	URL         string                 `json:"url,omitempty"`
	TemplateID  string                 `json:"template_id,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Severity    string                 `json:"severity,omitempty"`
	Message     string                 `json:"message"`
	Extractor   map[string]interface{} `json:"extractor,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	RawRequest  string                 `json:"raw_request,omitempty"`
	RawResponse string                 `json:"raw_response,omitempty"`
	CurlCommand string                 `json:"curl_command,omitempty"`
}

// Writer handles formatted output
type Writer struct {
	format   Format
	writer   io.Writer
	mu       sync.Mutex
	csvW     *csv.Writer
	firstRow bool
	results  []Result
}

// NewWriter creates a new output writer
func NewWriter(format Format, w io.Writer) *Writer {
	ow := &Writer{
		format:   format,
		writer:   w,
		firstRow: true,
	}
	if format == FormatCSV {
		ow.csvW = csv.NewWriter(w)
	}
	return ow
}

// WriteResult writes a single result
func (w *Writer) WriteResult(r Result) error {
	if w.format == FormatSilent {
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	switch w.format {
	case FormatJSON:
		return w.writeJSON(r)
	case FormatJSONL:
		return w.writeJSONL(r)
	case FormatCSV:
		return w.writeCSV(r)
	case FormatText:
		return w.writeText(r)
	default:
		return w.writeText(r)
	}
}

func (w *Writer) writeJSON(r Result) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w.writer, string(data))
	return err
}

func (w *Writer) writeJSONL(r Result) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w.writer, string(data))
	return err
}

func (w *Writer) writeCSV(r Result) error {
	if w.firstRow {
		w.csvW.Write([]string{"timestamp", "type", "level", "host", "port", "url", "template_id", "name", "severity", "message"})
		w.firstRow = false
	}
	w.csvW.Write([]string{
		r.Timestamp.Format(time.RFC3339),
		r.Type,
		r.Level,
		r.Host,
		fmt.Sprintf("%d", r.Port),
		r.URL,
		r.TemplateID,
		r.Name,
		r.Severity,
		r.Message,
	})
	w.csvW.Flush()
	return w.csvW.Error()
}

func (w *Writer) writeText(r Result) error {
	var color, reset string
	if f, ok := w.writer.(*os.File); ok && isTerminal(f) {
		switch strings.ToUpper(r.Level) {
		case "CRITICAL", "HIGH", "VUL", "VULNERABILITY":
			color = "\033[31m" // Red
		case "MEDIUM", "WARN":
			color = "\033[33m" // Yellow
		case "LOW", "INFO":
			color = "\033[32m" // Green
		case "DEBUG":
			color = "\033[36m" // Cyan
		default:
			color = "\033[37m" // White
		}
		reset = "\033[0m"
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("%s[%s]%s", color, strings.ToUpper(r.Level), reset))

	if r.Timestamp.IsZero() {
		r.Timestamp = time.Now()
	}
	parts = append(parts, r.Timestamp.Format("2006-01-02 15:04:05"))

	if r.TemplateID != "" {
		parts = append(parts, fmt.Sprintf("[%s]", r.TemplateID))
	}
	if r.Name != "" {
		parts = append(parts, r.Name)
	}
	if r.Host != "" && r.Port > 0 {
		parts = append(parts, fmt.Sprintf("%s:%d", r.Host, r.Port))
	}
	if r.URL != "" {
		parts = append(parts, r.URL)
	}
	if r.Message != "" {
		parts = append(parts, r.Message)
	}

	_, err := fmt.Fprintln(w.writer, strings.Join(parts, " "))
	return err
}

// Close flushes remaining output
func (w *Writer) Close() error {
	if w.csvW != nil {
		w.csvW.Flush()
	}
	return nil
}

// WriteString writes a raw string
func (w *Writer) WriteString(s string) error {
	_, err := fmt.Fprint(w.writer, s)
	return err
}

// isTerminal checks if file is a terminal
func isTerminal(f *os.File) bool {
	// Simplified check - on Windows this would use more complex logic
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return stat.Mode()&os.ModeCharDevice != 0
}

// ParseFormat parses a format string
func ParseFormat(s string) Format {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON
	case "jsonl", "ndjson":
		return FormatJSONL
	case "csv":
		return FormatCSV
	case "md", "markdown":
		return FormatMarkdown
	case "html":
		return FormatHTML
	case "silent", "none":
		return FormatSilent
	default:
		return FormatText
	}
}
