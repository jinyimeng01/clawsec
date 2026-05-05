package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/clawsec/clawsec/internal/logger"
)

// Server implements a built-in MCP server for ClawSec
type Server struct {
	port  int
	tools map[string]MCPTool
	http  *http.Server
}

// MCPTool represents an MCP tool definition
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	Handler     func(params map[string]interface{}) (interface{}, error)
}

// NewServer creates a new MCP server
func NewServer(port int) *Server {
	if port == 0 {
		port = 8080
	}

	s := &Server{
		port:  port,
		tools: make(map[string]MCPTool),
	}

	s.registerTools()

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp/tools", s.handleListTools)
	mux.HandleFunc("/mcp/call", s.handleCallTool)
	mux.HandleFunc("/mcp/health", s.handleHealth)

	s.http = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return s
}

func (s *Server) registerTools() {
	s.RegisterTool(MCPTool{
		Name:        "port_scan",
		Description: "Scan ports on a target host or CIDR range",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"target": map[string]string{"type": "string", "description": "Target IP or CIDR"},
				"ports":  map[string]string{"type": "string", "description": "Port range (e.g., 80,443,8080-8090)"},
			},
			"required": []string{"target"},
		},
		Handler: func(params map[string]interface{}) (interface{}, error) {
			target, _ := params["target"].(string)
			ports, _ := params["ports"].(string)
			return map[string]interface{}{
				"tool":    "port_scan",
				"target":  target,
				"ports":   ports,
				"command": fmt.Sprintf("clawsec scan port -t %s -p %s", target, ports),
			}, nil
		},
	})

	s.RegisterTool(MCPTool{
		Name:        "run_poc",
		Description: "Run a PoC template against a target",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"template": map[string]string{"type": "string", "description": "PoC template ID or path"},
				"target":   map[string]string{"type": "string", "description": "Target URL"},
			},
			"required": []string{"template", "target"},
		},
		Handler: func(params map[string]interface{}) (interface{}, error) {
			template, _ := params["template"].(string)
			target, _ := params["target"].(string)
			return map[string]interface{}{
				"tool":     "run_poc",
				"template": template,
				"target":   target,
				"command":  fmt.Sprintf("clawsec poc run -t %s -u %s --authorized", template, target),
			}, nil
		},
	})

	s.RegisterTool(MCPTool{
		Name:        "generate_report",
		Description: "Generate penetration test report",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"target": map[string]string{"type": "string", "description": "Target name"},
				"format": map[string]string{"type": "string", "description": "Report format (markdown, html, json)"},
			},
			"required": []string{"target"},
		},
		Handler: func(params map[string]interface{}) (interface{}, error) {
			target, _ := params["target"].(string)
			format, _ := params["format"].(string)
			if format == "" {
				format = "markdown"
			}
			return map[string]interface{}{
				"tool":    "generate_report",
				"target":  target,
				"format":  format,
				"command": fmt.Sprintf("clawsec ai report -t %s -o report.%s", target, format),
			}, nil
		},
	})
}

// RegisterTool registers a new tool
func (s *Server) RegisterTool(tool MCPTool) {
	s.tools[tool.Name] = tool
}

// Start starts the MCP server
func (s *Server) Start() error {
	logger.Infof("Starting MCP server on port %d", s.port)
	go func() {
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("MCP server error: %v", err)
		}
	}()
	return nil
}

// Stop stops the MCP server
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.http.Shutdown(ctx)
}

func (s *Server) handleListTools(w http.ResponseWriter, r *http.Request) {
	var toolList []map[string]interface{}
	for _, tool := range s.tools {
		toolList = append(toolList, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tools": toolList,
	})
}

func (s *Server) handleCallTool(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tool, ok := s.tools[req.Name]
	if !ok {
		http.Error(w, fmt.Sprintf("Tool not found: %s", req.Name), http.StatusNotFound)
		return
	}

	result, err := tool.Handler(req.Arguments)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"result": result,
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"tools":     len(s.tools),
		"timestamp": time.Now().Unix(),
	})
}
