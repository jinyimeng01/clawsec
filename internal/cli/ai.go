package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/clawsec/clawsec/internal/config"
	"github.com/clawsec/clawsec/internal/logger"
	"github.com/clawsec/clawsec/pkg/ai"
	"github.com/spf13/cobra"
)

var globalAgent *ai.Agent

func initAIAgent() (*ai.Agent, error) {
	if globalAgent != nil && globalAgent.IsReady() {
		return globalAgent, nil
	}

	logger.Infof("Initializing AI Agent...")
	agent, err := ai.NewAgent()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AI Agent: %w\n\nTo use AI features:\n1. Install Bun (https://bun.sh)\n2. Set ANTHROPIC_API_KEY environment variable\n3. Ensure ai-brain/ directory exists", err)
	}

	globalAgent = agent
	return agent, nil
}

func newAICommand() *cobra.Command {
	var (
		target   string
		context_ string
		_        string // mode reserved
		model    string
		apiKey   string
		endpoint string
	)

	aiCmd := &cobra.Command{
		Use:   "ai",
		Short: "AI security assistant - intelligent analysis and decision making",
		Long: `Interact with the AI security brain for intelligent offensive security analysis.

Powered by Anthropic Claude via Model Context Protocol (MCP).

Requires:
  - Bun runtime installed (https://bun.sh)
  - ANTHROPIC_API_KEY environment variable
  - ai-brain/ TypeScript agent built

Modes:
  analyze   - Analyze target and suggest attack paths
  suggest   - Suggest PoC templates based on fingerprints
  chain     - Build exploit chains from discovered vulnerabilities
  report    - Generate professional penetration test report
  chat      - Interactive AI security assistant

Examples:
  # Analyze a target and get attack recommendations
  clawsec ai analyze -t 10.0.0.1 --context "Apache 2.4.41, PHP 7.4, MySQL 5.7"

  # Suggest PoCs based on service fingerprint
  clawsec ai suggest -t http://target.com --fingerprint "Apache/2.4.41, PHP/7.4"

  # Generate report from scan results
  clawsec ai report -i results.json -o report.md

  # Interactive AI assistant
  clawsec ai chat`,
	}

	// Analyze subcommand
	analyzeCmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze target and suggest attack paths",
		RunE: func(cmd *cobra.Command, args []string) error {
			if target == "" {
				return fmt.Errorf("target is required")
			}

			agent, err := initAIAgent()
			if err != nil {
				return err
			}

			logger.Infof("AI analyzing target: %s", target)

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			var params map[string]interface{}
			if context_ != "" {
				params = map[string]interface{}{
					"target":   target,
					"services": parseContext(context_),
				}
			} else {
				params = map[string]interface{}{"target": target}
			}

			resp, err := agent.Call(ctx, "analyze", params)
			if err != nil {
				return fmt.Errorf("AI analysis failed: %w", err)
			}

			resultJSON, _ := json.MarshalIndent(resp.Result, "", "  ")
			fmt.Println(string(resultJSON))
			return nil
		},
	}
	analyzeCmd.Flags().StringVarP(&target, "target", "t", "", "target to analyze")
	analyzeCmd.Flags().StringVarP(&context_, "context", "c", "", "additional context (services, versions, etc.)")
	analyzeCmd.Flags().StringVar(&model, "model", "", "AI model to use")
	analyzeCmd.Flags().StringVar(&endpoint, "endpoint", "", "AI API endpoint")
	analyzeCmd.MarkFlagRequired("target")

	// Suggest subcommand
	suggestCmd := &cobra.Command{
		Use:   "suggest",
		Short: "Suggest PoC templates based on target fingerprints",
		RunE: func(cmd *cobra.Command, args []string) error {
			agent, err := initAIAgent()
			if err != nil {
				return err
			}

			logger.Infof("AI suggesting PoCs for: %s", target)

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			resp, err := agent.Call(ctx, "suggest", map[string]interface{}{
				"target":      target,
				"fingerprint": context_,
			})
			if err != nil {
				return fmt.Errorf("AI suggest failed: %w", err)
			}

			resultJSON, _ := json.MarshalIndent(resp.Result, "", "  ")
			fmt.Println(string(resultJSON))
			return nil
		},
	}
	suggestCmd.Flags().StringVarP(&target, "target", "t", "", "target URL")
	suggestCmd.Flags().StringVarP(&context_, "fingerprint", "f", "", "service fingerprint")
	suggestCmd.MarkFlagRequired("target")

	// Chain subcommand
	chainCmd := &cobra.Command{
		Use:   "chain",
		Short: "Build exploit chains from discovered vulnerabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			agent, err := initAIAgent()
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			resp, err := agent.Call(ctx, "chain", map[string]interface{}{
				"target":   target,
				"findings": []map[string]string{}, // TODO: load from file
			})
			if err != nil {
				return fmt.Errorf("AI chain failed: %w", err)
			}

			resultJSON, _ := json.MarshalIndent(resp.Result, "", "  ")
			fmt.Println(string(resultJSON))
			return nil
		},
	}
	chainCmd.Flags().StringVarP(&target, "target", "t", "", "target")
	chainCmd.Flags().StringVarP(&context_, "input", "i", "", "findings JSON file")
	chainCmd.MarkFlagRequired("target")

	// Report subcommand
	reportCmd := &cobra.Command{
		Use:   "report",
		Short: "Generate professional penetration test report",
		RunE: func(cmd *cobra.Command, args []string) error {
			agent, err := initAIAgent()
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()

			var findings []map[string]interface{}
			if context_ != "" {
				data, err := os.ReadFile(context_)
				if err != nil {
					return fmt.Errorf("failed to read findings file: %w", err)
				}
				json.Unmarshal(data, &findings)
			}

			resp, err := agent.Call(ctx, "report", map[string]interface{}{
				"target":   target,
				"findings": findings,
				"format":   "markdown",
			})
			if err != nil {
				return fmt.Errorf("AI report failed: %w", err)
			}

			// Extract report text
			resultMap, _ := resp.Result.(map[string]interface{})
			reportText, _ := resultMap["report"].(string)

			outFile := target
			if outFile == "" {
				outFile = "report.md"
			} else {
				outFile += "_report.md"
			}

			cfg := config.Get()
			if cfg.OutputFile != "" {
				outFile = cfg.OutputFile
			}

			if err := os.WriteFile(outFile, []byte(reportText), 0644); err != nil {
				return fmt.Errorf("failed to write report: %w", err)
			}

			logger.Infof("Report saved to: %s", outFile)
			return nil
		},
	}
	reportCmd.Flags().StringVarP(&target, "target", "t", "", "target name")
	reportCmd.Flags().StringVarP(&context_, "input", "i", "", "input findings file (JSON)")
	reportCmd.Flags().StringVarP(&target, "output", "o", "report.md", "output report file")

	// Chat subcommand
	chatCmd := &cobra.Command{
		Use:   "chat",
		Short: "Interactive AI security assistant",
		RunE: func(cmd *cobra.Command, args []string) error {
			agent, err := initAIAgent()
			if err != nil {
				return err
			}

			fmt.Println("╔════════════════════════════════════════════════════════════╗")
			fmt.Println("║           ClawSec AI Security Assistant                    ║")
			fmt.Println("║     Type 'exit' or 'quit' to end the session              ║")
			fmt.Println("╚════════════════════════════════════════════════════════════╝")
			fmt.Println()

			// If message provided as arg, send it once
			if len(args) > 0 {
				message := args[0]
				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				resp, err := agent.Call(ctx, "chat", map[string]interface{}{"message": message})
				cancel()
				if err != nil {
					return err
				}
				resultMap, _ := resp.Result.(map[string]interface{})
				response, _ := resultMap["response"].(string)
				fmt.Println("Assistant:", response)
				return nil
			}

			// Interactive mode
			for {
				fmt.Print("You: ")
				var input string
				fmt.Scanln(&input)

				if input == "exit" || input == "quit" {
					fmt.Println("Assistant: Goodbye!")
					break
				}

				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				resp, err := agent.Call(ctx, "chat", map[string]interface{}{"message": input})
				cancel()
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					continue
				}

				resultMap, _ := resp.Result.(map[string]interface{})
				response, _ := resultMap["response"].(string)
				fmt.Println("Assistant:", response)
				fmt.Println()
			}

			return nil
		},
	}
	chatCmd.Flags().StringVar(&model, "model", "", "AI model to use")
	chatCmd.Flags().StringVar(&apiKey, "api-key", "", "API key for AI service")
	chatCmd.Flags().StringVar(&endpoint, "endpoint", "", "AI API endpoint")

	aiCmd.AddCommand(analyzeCmd, suggestCmd, chainCmd, reportCmd, chatCmd)
	return aiCmd
}

func parseContext(ctx string) []map[string]string {
	// Simple parser for "product version, product version" format
	var services []map[string]string
	// TODO: implement proper parsing
	services = append(services, map[string]string{
		"product": ctx,
	})
	return services
}
