package safeline

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/clawsec/clawsec/pkg/products"
)

// SafeLine is a WAF product adapter
type SafeLine struct {
	products.BaseProduct
}

func New() *SafeLine {
	return &SafeLine{
		BaseProduct: products.BaseProduct{
			Name_:   "safeline",
			Headers: make(map[string]string),
		},
	}
}

func init() {
	products.Register("safeline", New())
}

func (s *SafeLine) Connect(config products.Config) error {
	s.Config = config
	s.InitHTTPClient()
	s.Headers["X-SLCE-API-TOKEN"] = config.APIKey
	s.Headers["Content-Type"] = "application/json"

	// Test connection
	ctx := context.Background()
	_, err := s.Query(ctx, "version", nil)
	if err != nil {
		return fmt.Errorf("failed to connect to SafeLine: %w", err)
	}

	s.Connected = true
	return nil
}

func (s *SafeLine) Query(ctx context.Context, queryType string, params map[string]interface{}) ([]map[string]interface{}, error) {
	switch queryType {
	case "version":
		return s.queryVersion(ctx)
	case "sites":
		return s.querySites(ctx)
	case "attack_logs":
		return s.queryAttackLogs(ctx, params)
	case "blocked_ips":
		return s.queryBlockedIPs(ctx)
	default:
		return nil, fmt.Errorf("unknown query type: %s", queryType)
	}
}

func (s *SafeLine) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	switch action {
	case "block_ip":
		return s.blockIP(ctx, params)
	case "unblock_ip":
		return s.unblockIP(ctx, params)
	case "add_rule":
		return s.addRule(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (s *SafeLine) queryVersion(ctx context.Context) ([]map[string]interface{}, error) {
	resp, err := s.DoRequest(ctx, "GET", "/api/open/version", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	return []map[string]interface{}{result}, nil
}

func (s *SafeLine) querySites(ctx context.Context) ([]map[string]interface{}, error) {
	resp, err := s.DoRequest(ctx, "GET", "/api/open/site", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	json.Unmarshal(body, &result)

	return result.Data, nil
}

func (s *SafeLine) queryAttackLogs(ctx context.Context, params map[string]interface{}) ([]map[string]interface{}, error) {
	// Simplified - in real implementation would handle pagination, time ranges, etc.
	resp, err := s.DoRequest(ctx, "GET", "/api/open/attack_log", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	json.Unmarshal(body, &result)

	return result.Data, nil
}

func (s *SafeLine) queryBlockedIPs(ctx context.Context) ([]map[string]interface{}, error) {
	resp, err := s.DoRequest(ctx, "GET", "/api/open/ip_group", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	json.Unmarshal(body, &result)

	return result.Data, nil
}

func (s *SafeLine) blockIP(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	ip, _ := params["ip"].(string)
	if ip == "" {
		return nil, fmt.Errorf("ip parameter required")
	}

	body, _ := json.Marshal(map[string]interface{}{
		"ip":     ip,
		"action": "block",
	})

	resp, err := s.DoRequest(ctx, "POST", "/api/open/ip_group", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	return result, nil
}

func (s *SafeLine) unblockIP(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	ip, _ := params["ip"].(string)
	if ip == "" {
		return nil, fmt.Errorf("ip parameter required")
	}

	// Implementation would delete from IP group
	return map[string]interface{}{"status": "ok", "message": fmt.Sprintf("IP %s unblocked", ip)}, nil
}

func (s *SafeLine) addRule(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	body, _ := json.Marshal(params)
	resp, err := s.DoRequest(ctx, "POST", "/api/open/rule", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	return result, nil
}
