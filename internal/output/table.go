package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"time"
)

// TableRenderer renders arbitrary data as formatted tables.
// Adapted from chaitin-cli's smart table renderer.
type TableRenderer struct {
	out io.Writer
}

// NewTableRenderer creates a new table renderer.
func NewTableRenderer(out io.Writer) *TableRenderer {
	return &TableRenderer{out: out}
}

// Render auto-detects data shape and renders accordingly.
func (r *TableRenderer) Render(data interface{}) error {
	extracted := extractData(data)
	if extracted == nil {
		fmt.Fprintln(r.out, "No data")
		return nil
	}

	val := reflect.ValueOf(extracted)
	switch val.Kind() {
	case reflect.Slice:
		return r.renderSlice(val)
	case reflect.Map:
		return r.renderMap(val)
	default:
		fmt.Fprintln(r.out, formatValueSimple(extracted))
	}
	return nil
}

func (r *TableRenderer) renderSlice(val reflect.Value) error {
	if val.Len() == 0 {
		fmt.Fprintln(r.out, "No data")
		return nil
	}

	first := val.Index(0).Interface()
	firstMap, ok := first.(map[string]interface{})
	if !ok {
		return r.renderSliceGeneric(val)
	}

	columns := selectColumns(firstMap)
	r.printHeader(columns)

	for i := 0; i < val.Len(); i++ {
		row := extractRowFromMap(val.Index(i).Interface(), columns)
		r.printRow(row)
	}
	return nil
}

func (r *TableRenderer) renderSliceGeneric(val reflect.Value) error {
	cols := inferColumns(val.Index(0).Interface())
	if len(cols) == 0 {
		fmt.Fprintln(r.out, "No data")
		return nil
	}
	r.printHeader(cols)
	for i := 0; i < val.Len(); i++ {
		r.printRow(extractRow(val.Index(i).Interface(), cols))
	}
	return nil
}

func (r *TableRenderer) renderMap(val reflect.Value) error {
	m, ok := val.Interface().(map[string]interface{})
	if !ok {
		fmt.Fprintln(r.out, formatValueSimple(val.Interface()))
		return nil
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	maxKeyLen := 0
	for _, k := range keys {
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
	}

	for _, k := range keys {
		fmt.Fprintf(r.out, "  %-*s  %s\n", maxKeyLen, strings.ToUpper(k), formatValueSimple(m[k]))
	}
	return nil
}

func (r *TableRenderer) printHeader(columns []string) {
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = strings.ToUpper(col)
	}
	seps := make([]string, len(columns))
	for i, h := range headers {
		seps[i] = strings.Repeat("-", max(len(h), 4))
	}
	r.printRow(headers)
	r.printRow(seps)
}

func (r *TableRenderer) printRow(row []string) {
	for i, col := range row {
		if i > 0 {
			fmt.Fprint(r.out, "  ")
		}
		fmt.Fprint(r.out, col)
	}
	fmt.Fprintln(r.out)
}

// selectColumns picks important columns first, then others, up to 10 total.
func selectColumns(m map[string]interface{}) []string {
	dataType := inferDataType(m)
	priority := importantFields[dataType]
	if priority == nil {
		priority = importantFields["default"]
	}

	var columns []string
	added := make(map[string]bool)

	for _, field := range priority {
		if _, exists := m[field]; exists {
			columns = append(columns, field)
			added[field] = true
		}
	}

	allFields := make([]string, 0, len(m))
	for k := range m {
		allFields = append(allFields, k)
	}
	sort.Strings(allFields)

	for _, field := range allFields {
		if added[field] {
			continue
		}
		if skipFields[field] {
			continue
		}
		columns = append(columns, field)
	}

	if len(columns) > 10 {
		columns = columns[:10]
	}
	return columns
}

var importantFields = map[string][]string{
	"default": {"id", "name", "status", "state", "enabled", "created_at", "updated_at"},
	"site":    {"id", "server_names", "ports", "upstreams", "state", "comment"},
	"rule":    {"id", "name", "is_enabled", "action", "pattern", "comment"},
	"ipgroup": {"id", "comment", "ips"},
	"attack":  {"id", "host", "url", "ip", "attack_type", "action", "created_at"},
	"audit":   {"id", "username", "content", "ip", "created_at"},
	"stat":    {"time", "value"},
	"product": {"name", "url", "status", "version"},
	"finding": {"name", "severity", "host", "url", "type"},
}

var skipFields = map[string]bool{
	"pattern": true, "auth_source_ids": true, "cloud_id": true, "cloud_total": true,
	"compatible": true, "builtin": true, "negate": true, "replay": true, "review": true,
	"tfa_enabled": true, "auth_callback": true, "auth_rule": true, "black_rule": true,
	"white_rule": true, "captcha_rule": true, "pass_count": true, "req_count": true,
	"expire": true, "level": true, "log": true, "init": true, "cert_id": true,
	"health_check": true, "exclude_paths": true, "forbidden_status_code": true,
	"not_found_status_code": true, "acl_response_html_path": true, "index": true,
	"load_balance": true, "redirect_status_code": true, "gateway_timeout_html_path": true,
	"gateway_timeout_status_code": true, "bad_gateway_html_path": true,
	"bad_gateway_status_code": true, "chaos_id": true, "chaos_is_enabled": true,
	"custom_location": true, "access_log_limit": true, "error_log_limit": true,
	"group_id": true, "email": true, "icon": true, "ssl": true, "stat_enabled": true,
	"sp_enabled": true, "static_default": true, "user_agent": true,
	"client_max_body_size": true, "cache": true, "cache_ttl": true, "retry": true,
	"owner": true, "state": true, "healthy": true, "health_state": true,
	"start_time": true, "end_time": true, "raw_request": true, "raw_response": true,
}

func inferDataType(m map[string]interface{}) string {
	if _, ok := m["server_names"]; ok {
		return "site"
	}
	if _, ok := m["pattern"]; ok {
		return "rule"
	}
	if _, ok := m["ips"]; ok {
		return "ipgroup"
	}
	if _, ok := m["attack_type"]; ok {
		return "attack"
	}
	if _, ok := m["severity"]; ok {
		return "finding"
	}
	return "default"
}

func extractData(data interface{}) interface{} {
	if data == nil {
		return nil
	}
	m, ok := data.(map[string]interface{})
	if !ok {
		return data
	}
	if d, exists := m["data"]; exists {
		return extractData(d)
	}
	return m
}

func extractRowFromMap(v interface{}, columns []string) []string {
	row := make([]string, len(columns))
	m, ok := v.(map[string]interface{})
	if !ok {
		return row
	}
	for i, col := range columns {
		if val, exists := m[col]; exists {
			row[i] = formatValueSimple(val)
		}
	}
	return row
}

func inferColumns(v interface{}) []string {
	if v == nil {
		return nil
	}
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct && val.Kind() != reflect.Map {
		return nil
	}

	var columns []string
	if val.Kind() == reflect.Map {
		for _, key := range val.MapKeys() {
			columns = append(columns, strings.ToUpper(key.String()))
		}
	} else {
		t := val.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" && jsonTag != "-" {
				name := strings.Split(jsonTag, ",")[0]
				if name != "" {
					columns = append(columns, strings.ToUpper(name))
				}
			} else {
				columns = append(columns, strings.ToUpper(field.Name))
			}
		}
	}
	return columns
}

func extractRow(v interface{}, columns []string) []string {
	row := make([]string, len(columns))
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	for i, col := range columns {
		row[i] = getFieldValue(val, strings.ToLower(col))
	}
	return row
}

func getFieldValue(val reflect.Value, fieldName string) string {
	if val.Kind() == reflect.Map {
		for _, key := range val.MapKeys() {
			if strings.ToLower(key.String()) == fieldName {
				return formatValueSimple(val.MapIndex(key).Interface())
			}
		}
		return ""
	}
	if val.Kind() == reflect.Struct {
		t := val.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			name := field.Name
			if jsonTag != "" && jsonTag != "-" {
				name = strings.Split(jsonTag, ",")[0]
			}
			if strings.ToLower(name) == fieldName {
				return formatValueSimple(val.Field(i).Interface())
			}
		}
	}
	return ""
}

func formatValueSimple(v interface{}) string {
	if v == nil {
		return ""
	}
	return formatValue(reflect.ValueOf(v))
}

func formatValue(v reflect.Value) string {
	if !v.IsValid() {
		return ""
	}
	if v.Kind() == reflect.Interface && !v.IsNil() {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if len(s) > 60 {
			return s[:57] + "..."
		}
		return s
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ts := v.Int()
		if ts > 1e12 && ts < 2e12 {
			return time.Unix(ts/1000, 0).Format("2006-01-02 15:04:05")
		} else if ts > 1e9 && ts < 2e9 {
			return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
		}
		return fmt.Sprintf("%d", ts)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		f := v.Float()
		if f > 1e12 && f < 2e12 {
			return time.Unix(int64(f/1000), 0).Format("2006-01-02 15:04:05")
		} else if f > 1e9 && f < 2e9 {
			return time.Unix(int64(f), 0).Format("2006-01-02 15:04:05")
		}
		if f == float64(int64(f)) {
			return fmt.Sprintf("%d", int64(f))
		}
		return fmt.Sprintf("%.2f", f)
	case reflect.Bool:
		if v.Bool() {
			return "✓"
		}
		return "✗"
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return "[]"
		}
		if v.Len() <= 3 {
			b, _ := json.Marshal(v.Interface())
			s := string(b)
			if len(s) <= 50 {
				return s
			}
		}
		return fmt.Sprintf("[%d items]", v.Len())
	case reflect.Map:
		if v.Len() == 0 {
			return "{}"
		}
		return fmt.Sprintf("{%d keys}", v.Len())
	case reflect.Struct:
		if t, ok := v.Interface().(time.Time); ok {
			return t.Format("2006-01-02 15:04:05")
		}
		b, _ := json.Marshal(v.Interface())
		s := string(b)
		if len(s) > 50 {
			return s[:47] + "..."
		}
		return s
	case reflect.Ptr:
		if v.IsNil() {
			return ""
		}
		return formatValue(v.Elem())
	default:
		s := fmt.Sprintf("%v", v.Interface())
		if len(s) > 60 {
			return s[:57] + "..."
		}
		return s
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
