package poc

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPClient performs HTTP requests for PoC execution
type HTTPClient struct {
	client         *http.Client
	timeout        time.Duration
	followRedirect bool
	maxRedirects   int
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(timeout int, followRedirect bool, maxRedirects int) *HTTPClient {
	if timeout <= 0 {
		timeout = 10
	}
	if maxRedirects <= 0 {
		maxRedirects = 10
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(timeout) * time.Second,
	}

	if !followRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return &HTTPClient{
		client:         client,
		timeout:        time.Duration(timeout) * time.Second,
		followRedirect: followRedirect,
		maxRedirects:   maxRedirects,
	}
}

// ExecuteHTTPRequest executes an HTTP request block
func (c *HTTPClient) ExecuteHTTPRequest(ctx context.Context, req HTTPRequest, baseURL string, vars map[string]interface{}) (*ResponseData, error) {
	// Build request
	path := req.Path[0]
	if len(req.Path) > 1 {
		// Multiple paths - for now just use first one
		path = req.Path[0]
	}

	// Expand variables in path
	path = ExpandVariables(path, vars)

	// Build full URL
	targetURL, err := buildURL(baseURL, path)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	// Determine method
	method := req.Method
	if method == "" {
		method = "GET"
	}

	// Build body
	body := req.Body
	body = ExpandVariables(body, vars)

	// Create HTTP request
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, targetURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range req.Headers {
		value = ExpandVariables(value, vars)
		httpReq.Header.Set(key, value)
	}

	// Set default headers if not present
	if httpReq.Header.Get("User-Agent") == "" {
		httpReq.Header.Set("User-Agent", "ClawSec/0.1.0")
	}
	if httpReq.Header.Get("Accept") == "" {
		httpReq.Header.Set("Accept", "*/*")
	}
	if body != "" && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// Execute request
	start := time.Now()
	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	duration := time.Since(start).Milliseconds()

	// Read body
	respBody, err := io.ReadAll(io.LimitReader(httpResp.Body, 10*1024*1024)) // 10MB limit
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Build headers string
	var headers strings.Builder
	headerMap := make(map[string]string)
	for key, values := range httpResp.Header {
		val := strings.Join(values, ", ")
		headers.WriteString(fmt.Sprintf("%s: %s\r\n", key, val))
		headerMap[key] = val
	}

	// Build raw response
	raw := fmt.Sprintf("HTTP/%d.%d %d %s\r\n%s\r\n%s",
		httpResp.ProtoMajor, httpResp.ProtoMinor,
		httpResp.StatusCode, httpResp.Status,
		headers.String(),
		string(respBody),
	)

	return &ResponseData{
		StatusCode: httpResp.StatusCode,
		Headers:    headers.String(),
		Body:       string(respBody),
		Raw:        raw,
		HeaderMap:  headerMap,
		Duration:   duration,
		Size:       len(respBody),
	}, nil
}

func buildURL(baseURL, path string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	// If path is absolute URL, use it directly
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path, nil
	}

	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	pathURL, err := url.Parse(path)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(pathURL).String(), nil
}
