package brute

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HTTPProtocol implements HTTP Basic/Digest brute force
type HTTPProtocol struct {
	Path     string
	Method   string
	Timeout  time.Duration
	Insecure bool
}

func (p *HTTPProtocol) Name() string {
	return "http"
}

func (p *HTTPProtocol) Try(ctx context.Context, target, username, password string) (Result, error) {
	if p.Timeout <= 0 {
		p.Timeout = 10 * time.Second
	}

	client := &http.Client{
		Timeout: p.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	url := target
	if p.Path != "" {
		url = target + p.Path
	}

	req, err := http.NewRequestWithContext(ctx, p.Method, url, nil)
	if err != nil {
		return Result{}, err
	}

	if p.Method == "" {
		req.Method = "GET"
	}

	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()

	result := Result{
		Target:   target,
		Protocol: "http",
		Username: username,
		Password: password,
	}

	// Check for successful authentication
	// 401 = Unauthorized (failed)
	// 200 or 302 = potentially successful
	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
		result.Success = true
		result.Banner = fmt.Sprintf("Status: %d", resp.StatusCode)
	}

	return result, nil
}
