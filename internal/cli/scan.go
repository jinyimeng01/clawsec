package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/clawsec/clawsec/internal/config"
	"github.com/clawsec/clawsec/internal/constants"
	"github.com/clawsec/clawsec/internal/logger"
	"github.com/clawsec/clawsec/internal/output"
	"github.com/clawsec/clawsec/pkg/engine/scanner"
	"github.com/clawsec/clawsec/internal/runner"
	"github.com/spf13/cobra"
)

func newScanCommand() *cobra.Command {
	var (
		targets    []string
		ports      string
		threads    int
		timeout    int
		_          string // scanType reserved
		rate       int
		bannerGrab bool
		osDetect   bool
	)

	scanCmd := &cobra.Command{
		Use:   "scan",
		Short: "Network scanning engine (port, service, asset discovery)",
		Long: `High-performance network scanning engine supporting multiple scan modes.

Scan Types:
  port      - TCP/UDP port scanning (SYN/Connect/UDP)
  service   - Service fingerprinting and version detection
  web       - Web asset discovery and technology fingerprinting
  host      - Host discovery (ICMP/ARP)

Examples:
  # TCP Connect scan of top 100 ports
  clawsec scan port -t 10.0.0.0/24

  # Full port SYN scan with banner grabbing
  clawsec scan port -t 10.0.0.1 -p 1-65535 --banner

  # Web fingerprinting
  clawsec scan web -t urls.txt

  # Service version detection
  clawsec scan service -t 10.0.0.1 -p 22,80,443,3306`,
	}

	// Port scan subcommand
	portCmd := &cobra.Command{
		Use:   "port",
		Short: "TCP/UDP port scanner",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(targets) == 0 {
				return fmt.Errorf("no targets specified, use -t flag")
			}

			cfg := config.Get()
			if threads == 0 {
				threads = cfg.Threads
			}
			if threads == 0 {
				threads = constants.DefaultThreads
			}
			if timeout == 0 {
				timeout = cfg.Timeout
			}
			if timeout == 0 {
				timeout = constants.DefaultScanTimeout
			}

			// Parse ports
			portList, err := scanner.ParsePorts(ports)
			if err != nil {
				return fmt.Errorf("invalid port specification: %w", err)
			}
			logger.Infof("Port range parsed: %d unique ports", len(portList))

			// Parse targets
			targetList, err := scanner.ParseTargets(targets, portList)
			if err != nil {
				return fmt.Errorf("invalid target specification: %w", err)
			}
			logger.Infof("Targets parsed: %d host/port combinations", len(targetList))

			if len(targetList) == 0 {
				return fmt.Errorf("no valid targets to scan")
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

			// Create scanner
			connectScanner := scanner.NewConnectScanner(threads, timeout, rate, bannerGrab)

			// Progress tracking
			progress := runner.NewProgress(int64(len(targetList)))
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			// Context for cancellation
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Start progress reporter
			done := make(chan struct{})
			go func() {
				defer close(done)
				for {
					select {
					case <-ticker.C:
						if !silent {
							logger.Infof(progress.String())
						}
					case <-ctx.Done():
						return
					}
				}
			}()

			// Run scan
			logger.Infof("Starting port scan - threads: %d, timeout: %ds, rate: %d/s, banner: %v",
				threads, timeout, rate, bannerGrab)
			startTime := time.Now()

			resultChan := connectScanner.Scan(ctx, targetList)
			var results []scanner.ScanResult

			for result := range resultChan {
				results = append(results, result)
				if result.Open {
					progress.IncrementOpen()
					addr := fmt.Sprintf("%s:%d", result.Target.IP.String(), result.Target.Port)
					msg := fmt.Sprintf("Port open: %d/%s", result.Target.Port, scanner.ProtoName(result.Target.Port))
					if result.Banner != "" {
						msg += fmt.Sprintf(" | Banner: %s", truncate(result.Banner, 80))
					}
					logger.Infof("[%s] %s (RTT: %v)", addr, msg, result.RTT)
				} else {
					progress.IncrementClosed()
				}
			}

			// Output results
			outResults := scanner.ResultsToOutput(results)
			for _, r := range outResults {
				out.WriteResult(r)
			}

			elapsed := time.Since(startTime)
			logger.Infof("Scan complete in %v", elapsed.Round(time.Millisecond))
			logger.Infof("Results: %d open / %d closed / %d total scanned",
				progress.Open, progress.Closed, progress.Completed)

			return nil
		},
	}
	portCmd.Flags().StringArrayVarP(&targets, "target", "t", nil, "target hosts/CIDR/URLs (required)")
	portCmd.Flags().StringVarP(&ports, "ports", "p", "top100", "port range (e.g., 80,443,8080-8090,top100,top1000)")
	portCmd.Flags().IntVar(&threads, "threads", 0, "concurrent threads (default: config or 50)")
	portCmd.Flags().IntVar(&timeout, "timeout", 0, "connection timeout in seconds (default: config or 3)")
	portCmd.Flags().IntVar(&rate, "rate", 0, "packets per second rate limit")
	portCmd.Flags().BoolVar(&bannerGrab, "banner", false, "grab service banners")
	portCmd.Flags().Bool("syn", false, "use SYN stealth scan (requires root)")
	portCmd.Flags().Bool("udp", false, "use UDP scan")
	portCmd.MarkFlagRequired("target")

	// Web scan subcommand
	webCmd := &cobra.Command{
		Use:   "web",
		Short: "Web asset scanner and fingerprinting",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("[INF] Web scanner - implementation in progress (Phase 2)")
			return nil
		},
	}
	webCmd.Flags().StringArrayVarP(&targets, "target", "t", nil, "target URLs/files")
	webCmd.Flags().IntVar(&threads, "threads", 30, "concurrent threads")
	webCmd.Flags().IntVar(&timeout, "timeout", 10, "request timeout in seconds")
	webCmd.MarkFlagRequired("target")

	// Service scan subcommand
	serviceCmd := &cobra.Command{
		Use:   "service",
		Short: "Service fingerprinting and version detection",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("[INF] Service fingerprinting - implementation in progress (Phase 2)")
			return nil
		},
	}
	serviceCmd.Flags().StringArrayVarP(&targets, "target", "t", nil, "target hosts")
	serviceCmd.Flags().StringVarP(&ports, "ports", "p", "top100", "ports to fingerprint")
	serviceCmd.Flags().BoolVar(&osDetect, "os-detect", false, "attempt OS detection")
	serviceCmd.MarkFlagRequired("target")

	scanCmd.AddCommand(portCmd, webCmd, serviceCmd)
	return scanCmd
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
