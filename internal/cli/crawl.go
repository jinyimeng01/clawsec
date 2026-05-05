package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/clawsec/clawsec/internal/config"
	"github.com/clawsec/clawsec/internal/logger"
	"github.com/clawsec/clawsec/pkg/engine/crawler"
	"github.com/spf13/cobra"
)

func newCrawlCommand() *cobra.Command {
	var (
		targets     []string
		wordlist    string
		threads     int
		timeout     int
		extensions  bool
		statusCodes []int
	)

	crawlCmd := &cobra.Command{
		Use:   "crawl",
		Short: "Web crawler and directory enumerator",
		Long: `Web crawling and directory enumeration engine.

Subcommands:
  dir         - Directory and file enumeration (dirbuster)
  js          - JavaScript file discovery and analysis
  params      - Parameter enumeration and fuzzing

Examples:
  # Directory enumeration with default wordlist
  clawsec crawl dir -t http://target.com

  # Directory enumeration with custom wordlist
  clawsec crawl dir -t http://target.com -w /path/to/wordlist.txt

  # With smart extensions and custom threads
  clawsec crawl dir -t http://target.com --ext -T 50`,
	}

	// Dir busting subcommand
	dirCmd := &cobra.Command{
		Use:   "dir",
		Short: "Directory and file enumeration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(targets) == 0 {
				return fmt.Errorf("no targets specified, use -t flag")
			}

			cfg := config.Get()
			if threads == 0 {
				threads = 20
			}
			if timeout == 0 {
				timeout = 10
			}

			// Load wordlist
			var words []string
			var err error
			if wordlist != "" {
				words, err = crawler.LoadWordlist(wordlist)
				if err != nil {
					return fmt.Errorf("failed to load wordlist: %w", err)
				}
			} else {
				words = crawler.DefaultWordlist()
			}

			if extensions {
				words = crawler.SmartExtensions(words)
				logger.Infof("Smart extensions enabled: %d total paths", len(words))
			}

			// Setup output
			outFile := cfg.OutputFile
			var outWriter = os.Stdout
			if outFile != "" {
				f, err := os.Create(outFile)
				if err != nil {
					return fmt.Errorf("failed to create output file: %w", err)
				}
				defer f.Close()
				outWriter = f
			}

			logger.Infof("Starting directory enumeration - targets: %d, wordlist: %d, threads: %d",
				len(targets), len(words), threads)

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout*len(words))*time.Second)
			defer cancel()

			foundCount := 0
			for _, target := range targets {
				db := crawler.NewDirBuster(threads, timeout)
				if len(statusCodes) > 0 {
					db.SetStatusFilter(statusCodes)
				}

				start := time.Now()
				results := db.Scan(ctx, target, words)

				for result := range results {
					foundCount++
					line := fmt.Sprintf("[%d] %s", result.StatusCode, result.URL)
					if result.Title != "" {
						line += fmt.Sprintf(" | Title: %s", result.Title)
					}
					if result.Redirect != "" {
						line += fmt.Sprintf(" -> %s", result.Redirect)
					}
					if result.Size > 0 {
						line += fmt.Sprintf(" (%d bytes)", result.Size)
					}
					fmt.Fprintln(outWriter, line)
					logger.Infof("Found: %s", line)
				}

				logger.Infof("Target %s completed in %v", target, time.Since(start).Round(time.Millisecond))
			}

			logger.Infof("Enumeration complete. Found %d entries.", foundCount)
			return nil
		},
	}
	dirCmd.Flags().StringArrayVarP(&targets, "target", "t", nil, "target URLs (required)")
	dirCmd.Flags().StringVarP(&wordlist, "wordlist", "w", "", "wordlist file path")
	dirCmd.Flags().IntVarP(&threads, "threads", "T", 20, "concurrent threads")
	dirCmd.Flags().IntVar(&timeout, "timeout", 10, "request timeout in seconds")
	dirCmd.Flags().BoolVar(&extensions, "ext", false, "add smart extensions (.php, .bak, .zip, etc.)")
	dirCmd.Flags().IntSliceVar(&statusCodes, "status", nil, "filter by status codes (default: 200,201,204,301,302,307,308,401,403,405,500)")
	dirCmd.MarkFlagRequired("target")

	// JS discovery subcommand (placeholder)
	jsCmd := &cobra.Command{
		Use:   "js",
		Short: "JavaScript file discovery and endpoint extraction",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("[INF] JavaScript discovery - implementation in progress")
			return nil
		},
	}
	jsCmd.Flags().StringArrayVarP(&targets, "target", "t", nil, "target URLs")
	jsCmd.MarkFlagRequired("target")

	// Parameter fuzzing subcommand (placeholder)
	paramCmd := &cobra.Command{
		Use:   "params",
		Short: "Parameter enumeration and fuzzing",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("[INF] Parameter fuzzing - implementation in progress")
			return nil
		},
	}
	paramCmd.Flags().StringArrayVarP(&targets, "target", "t", nil, "target URLs")
	paramCmd.MarkFlagRequired("target")

	crawlCmd.AddCommand(dirCmd, jsCmd, paramCmd)
	return crawlCmd
}
