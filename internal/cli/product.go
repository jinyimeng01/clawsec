package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/clawsec/clawsec/internal/config"
	"github.com/clawsec/clawsec/pkg/products"
	"github.com/spf13/cobra"
)

func newProductCommand() *cobra.Command {
	productCmd := &cobra.Command{
		Use:   "product",
		Short: "Unified security product management console",
		Long: `Manage and interact with various security products from a unified CLI interface.

Supported Products:
  safeline      - Chaitin SafeLine WAF
  xray          - Chaitin X-Ray vulnerability scanner
  cloudwalker   - Chaitin CloudWalker CWPP
  tanswer       - Chaitin T-Answer traffic threat detection
  ddr           - Chaitin DDR data security

Commands:
  list          - List configured products
  config        - Configure product credentials
  query         - Query product data
  exec          - Execute product commands

Examples:
  # List configured products
  clawsec product list

  # Query WAF attack logs
  clawsec product query safeline attack_logs

  # Block IP on WAF
  clawsec product exec safeline block_ip --ip 1.2.3.4`,
	}

	// List subcommand
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List configured security products",
		RunE: func(cmd *cobra.Command, args []string) error {
			registered := products.List()
			cfg := config.Get()

			fmt.Println("Registered products:")
			for _, name := range registered {
				if pc, ok := cfg.GetProduct(name); ok && pc.URL != "" {
					fmt.Printf("  ✓ %-15s [%s]\n", name, pc.URL)
				} else {
					fmt.Printf("  ✗ %-15s [not configured]\n", name)
				}
			}
			return nil
		},
	}

	// Config subcommand
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configure product credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("product name required")
			}
			productName := args[0]

			// Interactive config (simplified)
			fmt.Printf("Configuring %s...\n", productName)
			fmt.Print("URL: ")
			var url string
			fmt.Scanln(&url)
			fmt.Print("API Key: ")
			var apiKey string
			fmt.Scanln(&apiKey)

			cfg := config.Get()
			cfg.SetProduct(productName, config.ProductConfig{
				URL:    url,
				APIKey: apiKey,
			})

			fmt.Printf("%s configured.\n", productName)
			return nil
		},
	}

	// Query subcommand
	queryCmd := &cobra.Command{
		Use:   "query",
		Short: "Query product data",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: clawsec product query <product> <query-type>")
			}

			productName := args[0]
			queryType := args[1]

			p, ok := products.Get(productName)
			if !ok {
				return fmt.Errorf("unknown product: %s", productName)
			}

			cfg := config.Get()
			pc, configured := cfg.GetProduct(productName)
			if !configured || pc.URL == "" {
				return fmt.Errorf("product %s not configured. Run 'clawsec product config %s' first", productName, productName)
			}

			if err := p.Connect(products.Config{
				URL:      pc.URL,
				APIKey:   pc.APIKey,
				Token:    pc.Token,
				Insecure: pc.Insecure,
			}); err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			results, err := p.Query(ctx, queryType, nil)
			if err != nil {
				return fmt.Errorf("query failed: %w", err)
			}

			output, _ := json.MarshalIndent(results, "", "  ")
			fmt.Println(string(output))
			return nil
		},
	}

	// Exec subcommand
	execCmd := &cobra.Command{
		Use:   "exec",
		Short: "Execute product commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: clawsec product exec <product> <action>")
			}

			productName := args[0]
			action := args[1]

			p, ok := products.Get(productName)
			if !ok {
				return fmt.Errorf("unknown product: %s", productName)
			}

			cfg := config.Get()
			pc, configured := cfg.GetProduct(productName)
			if !configured || pc.URL == "" {
				return fmt.Errorf("product %s not configured", productName)
			}

			if err := p.Connect(products.Config{
				URL:      pc.URL,
				APIKey:   pc.APIKey,
				Token:    pc.Token,
				Insecure: pc.Insecure,
			}); err != nil {
				return err
			}

			// Parse remaining args as key=value params
			params := make(map[string]interface{})
			for _, arg := range args[2:] {
				if idx := len(arg); idx > 0 {
					for i, c := range arg {
						if c == '=' {
							params[arg[:i]] = arg[i+1:]
							break
						}
					}
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			result, err := p.Execute(ctx, action, params)
			if err != nil {
				return fmt.Errorf("execution failed: %w", err)
			}

			output, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(output))
			return nil
		},
	}

	productCmd.AddCommand(listCmd, configCmd, queryCmd, execCmd)
	return productCmd
}
