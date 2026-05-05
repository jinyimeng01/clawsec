package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/clawsec/clawsec/internal/config"
	"github.com/clawsec/clawsec/internal/logger"
	"github.com/clawsec/clawsec/internal/output"
	"github.com/clawsec/clawsec/pkg/engine/poc"
	"github.com/spf13/cobra"
)

func newPoCCommand() *cobra.Command {
	var (
		targetURLs  []string
		templates   []string
		templateDir string
		severity    []string
		tags        []string
		workflow    string
		_           bool // updateTemplates reserved
		_           bool // listTemplates reserved
		stats       bool
		threads     int
		bulkSize    int
		retries     int
		timeout     int
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
  clawsec poc run -d ./poc/ -u targets.txt

  # Run templates filtered by severity and tags
  clawsec poc run -d ./nuclei-templates/ -u targets.txt -s critical,high -t cve,rce

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
				return fmt.Errorf("no templates specified, use -t or -d flag")
			}

			cfg := config.Get()

			// Load templates
			var templatePaths []string
			for _, t := range templates {
				if info, err := os.Stat(t); err == nil && info.IsDir() {
					templatePaths = append(templatePaths, t)
				} else {
					templatePaths = append(templatePaths, t)
				}
			}
			if templateDir != "" {
				templatePaths = append(templatePaths, templateDir)
			}

			logger.Infof("Loading templates from %d paths...", len(templatePaths))
			loadedTemplates, err := poc.LoadTemplatesFromPaths(templatePaths)
			if err != nil {
				return fmt.Errorf("failed to load templates: %w", err)
			}
			logger.Infof("Loaded %d templates", len(loadedTemplates))

			// Filter templates
			loadedTemplates = poc.FilterTemplates(loadedTemplates, severity, tags)
			logger.Infof("After filtering: %d templates", len(loadedTemplates))

			if len(loadedTemplates) == 0 {
				return fmt.Errorf("no templates match the specified criteria")
			}

			// Setup output
			var outWriter = os.Stdout
			if cfg.OutputFile != "" {
				f, err := os.Create(cfg.OutputFile)
				if err != nil {
					return fmt.Errorf("failed to create output file: %w", err)
				}
				defer f.Close()
				outWriter = f
			}
			out := output.NewWriter(output.ParseFormat(cfg.OutputFormat), outWriter)
			defer out.Close()

			// Execute templates
			if threads == 0 {
				threads = 25
			}
			executor := poc.NewExecutor(threads, timeout)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()

			logger.Infof("Starting PoC execution - templates: %d, targets: %d, threads: %d",
				len(loadedTemplates), len(targetURLs), threads)
			startTime := time.Now()

			matchedCount := 0
			for result := range executor.ExecuteMultiple(ctx, loadedTemplates, targetURLs) {
				matchedCount++
				logger.Infof("[VUL] %s - %s (%s) - %s",
					result.TemplateID,
					result.Info.Name,
					result.Info.Severity,
					result.URL)

				// Output result
				out.WriteResult(output.Result{
					Timestamp:  result.MatchedAt,
					Type:       "vulnerability",
					Level:      result.Info.Severity,
					Host:       result.Host,
					URL:        result.URL,
					TemplateID: result.TemplateID,
					Name:       result.Info.Name,
					Severity:   result.Info.Severity,
					Message:    fmt.Sprintf("%s - %s", result.TemplateID, result.Info.Name),
					Metadata:   result.Meta,
				})
			}

			elapsed := time.Since(startTime)
			logger.Infof("PoC execution complete in %v", elapsed.Round(time.Millisecond))
			logger.Infof("Results: %d vulnerabilities found", matchedCount)

			return nil
		},
	}
	runCmd.Flags().StringArrayVarP(&targetURLs, "url", "u", nil, "target URLs/files (required)")
	runCmd.Flags().StringArrayVarP(&templates, "template", "t", nil, "template files")
	runCmd.Flags().StringVarP(&templateDir, "template-dir", "d", "", "template directory")
	runCmd.Flags().StringArrayVar(&severity, "severity", nil, "filter by severity (critical,high,medium,low,info)")
	runCmd.Flags().StringArrayVar(&tags, "tags", nil, "filter by tags")
	runCmd.Flags().StringVarP(&workflow, "workflow", "w", "", "workflow file for chained execution")
	runCmd.Flags().IntVar(&threads, "threads", 25, "concurrent threads")
	runCmd.Flags().IntVar(&bulkSize, "bulk-size", 25, "bulk size for templates")
	runCmd.Flags().IntVar(&retries, "retries", 1, "number of retries")
	runCmd.Flags().IntVar(&timeout, "timeout", 10, "request timeout in seconds")
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
	listCmd.Flags().StringArrayVar(&severity, "severity", nil, "filter by severity")
	listCmd.Flags().StringArrayVar(&tags, "tags", nil, "filter by tags")

	pocCmd.AddCommand(runCmd, updateCmd, listCmd)
	return pocCmd
}

// isTemplate checks if file is a valid template
func isTemplate(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}
