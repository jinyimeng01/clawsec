package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newPoCCommand() *cobra.Command {
	var (
		targetURLs   []string
		templates    []string
		templateDir  string
		severity     []string
		tags         []string
		workflow     string
		_ bool // updateTemplates reserved
		_ bool // listTemplates reserved
		stats        bool
		threads      int
		bulkSize     int
		retries      int
	)

	pocCmd := &cobra.Command{
		Use:   "poc",
		Short: "PoC engine - Nuclei-compatible vulnerability verification",
		Long: `Execute vulnerability proof-of-concept templates compatible with Nuclei YAML format.

Features:
  • Full Nuclei YAML template syntax support
  • HTTP/TCP/UDP/DNS/SSL/WebSocket/Headless/Code protocols
  • DSL expression engine with 50+ built-in functions
  • Multi-step workflow chains with variable passing
  • Automatic template updates from community repository

Examples:
  # Run a single template against a target
  clawsec poc run -t CVE-2021-41773.yaml -u http://target.com

  # Run all templates in a directory against multiple targets
  clawsec poc run -td ./poc/ -u targets.txt

  # Run templates filtered by severity and tags
  clawsec poc run -td ./nuclei-templates/ -u targets.txt -s critical,high -t cve,rce

  # Update templates from remote repository
  clawsec poc update

  # List available templates
  clawsec poc list`,
	}

	// Run subcommand
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run PoC templates against targets",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(targetURLs) == 0 {
				return fmt.Errorf("no target URLs specified, use -u flag")
			}
			if len(templates) == 0 && templateDir == "" {
				return fmt.Errorf("no templates specified, use -t or -td flag")
			}

			fmt.Printf("[INF] PoC Engine initializing...\n")
			fmt.Printf("[INF] Targets: %d | Templates: %v | Threads: %d\n",
				len(targetURLs), templates, threads)
			fmt.Println("[INF] PoC engine - implementation in progress (Phase 3)")
			return nil
		},
	}
	runCmd.Flags().StringArrayVarP(&targetURLs, "url", "u", nil, "target URLs/files (required)")
	runCmd.Flags().StringArrayVarP(&templates, "template", "t", nil, "template files")
	runCmd.Flags().StringVarP(&templateDir, "template-dir", "d", "", "template directory")
	runCmd.Flags().StringArrayVarP(&severity, "severity", "s", nil, "filter by severity (critical,high,medium,low,info)")
	runCmd.Flags().StringArrayVar(&tags, "tags", nil, "filter by tags")
	runCmd.Flags().StringVarP(&workflow, "workflow", "w", "", "workflow file for chained execution")
	runCmd.Flags().IntVar(&threads, "threads", 25, "concurrent threads")
	runCmd.Flags().IntVar(&bulkSize, "bulk-size", 25, "bulk size for templates")
	runCmd.Flags().IntVar(&retries, "retries", 1, "number of retries")
	runCmd.MarkFlagRequired("url")

	// Update subcommand
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update PoC templates from remote repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("[INF] Checking for template updates...")
			fmt.Println("[INF] Template updater - implementation in progress (Phase 3)")
			return nil
		},
	}

	// List subcommand
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available PoC templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := templateDir
			if dir == "" {
				dir = filepath.Join("templates", "poc")
			}
			fmt.Printf("[INF] Listing templates in: %s\n", dir)
			fmt.Println("[INF] Template lister - implementation in progress (Phase 3)")
			return nil
		},
	}
	listCmd.Flags().StringVarP(&templateDir, "template-dir", "d", "", "template directory")
	listCmd.Flags().BoolVar(&stats, "stats", false, "show template statistics")
	listCmd.Flags().StringArrayVarP(&severity, "severity", "s", nil, "filter by severity")
	listCmd.Flags().StringArrayVar(&tags, "tags", nil, "filter by tags")

	pocCmd.AddCommand(runCmd, updateCmd, listCmd)
	return pocCmd
}

// isTemplate checks if file is a valid template
func isTemplate(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}
