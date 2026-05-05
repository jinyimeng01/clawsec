package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/clawsec/clawsec/internal/logger"
)

// Agent manages the AI Brain subprocess
type Agent struct {
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	stdout   io.ReadCloser
	mu       sync.Mutex
	ready    bool
	requests map[string]chan *Response
	running  bool
}

// NewAgent creates and starts the AI Agent subprocess
func NewAgent() (*Agent, error) {
	agent := &Agent{
		requests: make(map[string]chan *Response),
	}

	if err := agent.start(); err != nil {
		return nil, err
	}

	go agent.readLoop()

	// Wait for agent to be ready
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := agent.waitReady(ctx); err != nil {
		agent.Stop()
		return nil, fmt.Errorf("AI Agent failed to start: %w", err)
	}

	return agent, nil
}

func (a *Agent) start() error {
	// Find ai-brain entry point
	brainPath := a.findBrainPath()
	if brainPath == "" {
		return fmt.Errorf("AI Brain not found. Please ensure ai-brain/ is built and available")
	}

	logger.Infof("Starting AI Brain from: %s", brainPath)

	var cmd *exec.Cmd
	if filepath.Ext(brainPath) == ".ts" {
		// Run with bun
		cmd = exec.Command("bun", "run", brainPath)
	} else if runtime.GOOS == "windows" && filepath.Ext(brainPath) == ".exe" {
		cmd = exec.Command(brainPath)
	} else {
		cmd = exec.Command(brainPath)
	}

	// Inherit environment
	cmd.Env = os.Environ()

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start AI Brain: %w", err)
	}

	a.cmd = cmd
	a.stdin = stdin
	a.stdout = stdout
	a.running = true

	return nil
}

func (a *Agent) findBrainPath() string {
	candidates := []string{
		"ai-brain/src/main.ts",
		"ai-brain/dist/main.js",
		"ai-brain/dist/main",
		"../ai-brain/src/main.ts",
		"../ai-brain/dist/main.js",
	}

	// Also check relative to executable
	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "ai-brain", "src", "main.ts"),
			filepath.Join(exeDir, "ai-brain", "dist", "main.js"),
			filepath.Join(exeDir, "..", "ai-brain", "src", "main.ts"),
		)
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

func (a *Agent) waitReady(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		resp, err := a.Call(ctx, "ping", nil)
		if err == nil && resp != nil {
			logger.Infof("AI Brain ready: %v", resp)
			a.ready = true
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func (a *Agent) readLoop() {
	scanner := bufio.NewScanner(a.stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var resp Response
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			logger.Debugf("AI Brain non-JSON output: %s", line)
			continue
		}

		a.mu.Lock()
		ch, ok := a.requests[resp.ID]
		a.mu.Unlock()

		if ok && ch != nil {
			ch <- &resp
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Debugf("AI Brain read error: %v", err)
	}

	a.mu.Lock()
	a.running = false
	a.ready = false
	a.mu.Unlock()
}

// Call sends a JSON-RPC request and waits for response
func (a *Agent) Call(ctx context.Context, method string, params interface{}) (*Response, error) {
	a.mu.Lock()
	if !a.running {
		a.mu.Unlock()
		return nil, fmt.Errorf("AI Agent not running")
	}

	id := fmt.Sprintf("req_%d", time.Now().UnixNano())
	ch := make(chan *Response, 1)
	a.requests[id] = ch
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		delete(a.requests, id)
		a.mu.Unlock()
	}()

	req := Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	if _, err := fmt.Fprintln(a.stdin, string(data)); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-ch:
		if resp.Error != nil {
			return nil, fmt.Errorf("AI Agent error: %s", resp.Error.Message)
		}
		return resp, nil
	}
}

// Stop terminates the AI Agent subprocess
func (a *Agent) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return
	}

	a.running = false
	a.ready = false

	if a.stdin != nil {
		a.stdin.Close()
	}

	if a.cmd != nil && a.cmd.Process != nil {
		a.cmd.Process.Kill()
		a.cmd.Wait()
	}

	logger.Infof("AI Brain stopped")
}

// IsReady returns true if agent is ready
func (a *Agent) IsReady() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.ready
}
