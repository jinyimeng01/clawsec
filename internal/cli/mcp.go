package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/clawsec/clawsec/internal/logger"
	"github.com/clawsec/clawsec/pkg/mcp"
	"github.com/spf13/cobra"
)

func newMCPServeCommand() *cobra.Command {
	var port int

	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP (Model Context Protocol) server management",
		Long: `Run ClawSec as an MCP server to expose security tools to AI agents.

The MCP server provides HTTP endpoints for external AI systems to:
- Query available security tools
- Execute port scans, PoC runs, and report generation
- Integrate with Claude, Cursor, and other MCP-compatible clients

Endpoints:
  GET  /mcp/health   - Health check
  GET  /mcp/tools    - List available tools
  POST /mcp/call     - Execute a tool

Example:
  clawsec mcp serve --port 8080`,
	}

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start MCP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			server := mcp.NewServer(port)
			if err := server.Start(); err != nil {
				return fmt.Errorf("failed to start MCP server: %w", err)
			}

			logger.Infof("MCP Server running on http://localhost:%d", port)
			logger.Infof("Health check: curl http://localhost:%d/mcp/health", port)
			logger.Infof("Press Ctrl+C to stop")

			// Wait for interrupt
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			<-sigChan

			logger.Infof("Shutting down MCP server...")
			server.Stop()
			return nil
		},
	}

	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "MCP server port")
	mcpCmd.AddCommand(serveCmd)
	return mcpCmd
}
