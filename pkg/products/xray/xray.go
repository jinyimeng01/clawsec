package xray

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/clawsec/clawsec/pkg/products"
)

// XRay is a vulnerability scanner adapter
type XRay struct {
	products.BaseProduct
}

func New() *XRay {
	return &XRay{
		BaseProduct: products.BaseProduct{
			Name_:   "xray",
			Headers: make(map[string]string),
		},
	}
}

func init() {
	products.Register("xray", New())
}

func (x *XRay) Connect(config products.Config) error {
	x.Config = config
	x.InitHTTPClient()
	x.Headers["Authorization"] = "Token " + config.APIKey
	x.Headers["Content-Type"] = "application/json"

	ctx := context.Background()
	_, err := x.Query(ctx, "version", nil)
	if err != nil {
		return fmt.Errorf("failed to connect to X-Ray: %w", err)
	}

	x.Connected = true
	return nil
}

func (x *XRay) Query(ctx context.Context, queryType string, params map[string]interface{}) ([]map[string]interface{}, error) {
	switch queryType {
	case "version":
		return x.queryVersion(ctx)
	case "tasks":
		return x.queryTasks(ctx)
	case "vulnerabilities":
		return x.queryVulnerabilities(ctx, params)
	case "assets":
		return x.queryAssets(ctx)
	default:
		return nil, fmt.Errorf("unknown query type: %s", queryType)
	}
}

func (x *XRay) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	switch action {
	case "create_task":
		return x.createTask(ctx, params)
	case "start_task":
		return x.startTask(ctx, params)
	case "stop_task":
		return x.stopTask(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (x *XRay) queryVersion(ctx context.Context) ([]map[string]interface{}, error) {
	resp, err := x.DoRequest(ctx, "GET", "/api/v1/system/version", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	return []map[string]interface{}{result}, nil
}

func (x *XRay) queryTasks(ctx context.Context) ([]map[string]interface{}, error) {
	resp, err := x.DoRequest(ctx, "GET", "/api/v1/task", nil)
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

func (x *XRay) queryVulnerabilities(ctx context.Context, params map[string]interface{}) ([]map[string]interface{}, error) {
	resp, err := x.DoRequest(ctx, "GET", "/api/v1/vulnerability", nil)
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

func (x *XRay) queryAssets(ctx context.Context) ([]map[string]interface{}, error) {
	resp, err := x.DoRequest(ctx, "GET", "/api/v1/asset", nil)
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

func (x *XRay) createTask(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	body, _ := json.Marshal(params)
	resp, err := x.DoRequest(ctx, "POST", "/api/v1/task", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	return result, nil
}

func (x *XRay) startTask(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	taskID, _ := params["id"].(string)
	if taskID == "" {
		return nil, fmt.Errorf("task id required")
	}

	body, _ := json.Marshal(map[string]interface{}{"status": "running"})
	resp, err := x.DoRequest(ctx, "PUT", "/api/v1/task/"+taskID, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	return result, nil
}

func (x *XRay) stopTask(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	taskID, _ := params["id"].(string)
	if taskID == "" {
		return nil, fmt.Errorf("task id required")
	}

	body, _ := json.Marshal(map[string]interface{}{"status": "stopped"})
	resp, err := x.DoRequest(ctx, "PUT", "/api/v1/task/"+taskID, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	return result, nil
}
