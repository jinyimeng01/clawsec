package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/clawsec/clawsec/internal/config"
	"github.com/clawsec/clawsec/internal/constants"
	"github.com/clawsec/clawsec/internal/logger"
	"github.com/clawsec/clawsec/internal/output"
	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	outputFile string
	outputFmt  string
	silent     bool
	verbose    bool
	debug      bool
	noColor    bool
	authorized bool
)

// NewRootCommand creates the root command
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   constants.AppSlug,
		Short: fmt.Sprintf("%s - Global Top-tier AI-Driven Offensive Security CLI Platform", constants.AppName),
		Long: fmt.Sprintf(`%s v%s

A unified AI-Native network offensive security testing platform.

Core Capabilities:
  • Port Scanning    - SYN/Connect/UDP scanning with adaptive rate control
  • PoC Engine       - Nuclei YAML compatible vulnerability verification
  • Brute Force      - 10+ protocol password brute-forcing
  • Web Fuzzing      - Directory enumeration, JS extraction, parameter fuzzing
  • AI Agent         - Intelligent target analysis and exploit chain building
  • Product Console  - Unified management for WAF/Scanner/EDR/CMDB

For more information: https://github.com/clawsec/clawsec`, constants.AppName, constants.Version),
		Version:           fmt.Sprintf("%s (build: %s, commit: %s)", constants.Version, constants.BuildTime, constants.GitCommit),
		DisableAutoGenTag: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initialize(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Global flags
	root.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path (default: ~/.clawsec/config.yaml)")
	root.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "output file (default: stdout)")
	root.PersistentFlags().StringVarP(&outputFmt, "format", "f", "text", "output format: text/json/jsonl/csv")
	root.PersistentFlags().BoolVarP(&silent, "silent", "s", false, "silent mode (no output)")
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	root.PersistentFlags().BoolVar(&debug, "debug", false, "debug mode (maximum verbosity)")
	root.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	root.PersistentFlags().BoolVar(&authorized, "authorized", false, "I am authorized to perform offensive security testing on the target")
	root.PersistentFlags().Bool("ai", false, "Enable AI assistant for intelligent analysis")

	// Add subcommands
	root.AddCommand(newScanCommand())
	root.AddCommand(newPoCCommand())
	root.AddCommand(newBruteCommand())
	root.AddCommand(newCrawlCommand())
	root.AddCommand(newAICommand())
	root.AddCommand(newProductCommand())
	root.AddCommand(newWorkflowCommand())
	root.AddCommand(newMCPServeCommand())
	root.AddCommand(newVersionCommand())

	// Cleanup on exit
	cobra.OnFinalize(func() {
		if globalAgent != nil {
			globalAgent.Stop()
		}
	})

	return root
}

func initialize(cmd *cobra.Command) error {
	// Determine log level
	level := logger.InfoLevel
	if debug {
		level = logger.DebugLevel
		verbose = true
	}

	// Initialize logger
	logOpts := []logger.Option{
		logger.WithLevel(level),
		logger.WithColor(!noColor && !silent),
	}
	if debug {
		logOpts = append(logOpts, logger.WithJSON(true))
	}
	logger.Init(logOpts...)

	// Load config
	cfg := config.Get()
	cfg.Silent = silent
	cfg.Verbose = verbose
	cfg.Debug = debug
	cfg.NoColor = noColor
	cfg.Authorized = authorized
	cfg.OutputFormat = outputFmt
	cfg.OutputFile = outputFile

	// Config file path resolution
	if cfgFile == "" {
		cfgFile = config.DefaultConfigPath()
	}
	cfg.ConfigPath = cfgFile

	// Load config file (YAML + .env + env vars with auto-reflection)
	if err := cfg.LoadFile(cfgFile); err != nil {
		logger.Warnf("Failed to load config: %v", err)
	}

	// Ensure default config exists
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		if err := config.InitDefaultConfig(); err != nil {
			logger.Warnf("Failed to create default config: %v", err)
		}
	}

	// Setup output writer
	var outWriter = os.Stdout
	if outputFile != "" {
		dir := filepath.Dir(outputFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		f, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		outWriter = f
	}

	if silent {
		outWriter, _ = os.Open(os.DevNull)
	}

	_ = output.NewWriter(output.ParseFormat(outputFmt), outWriter)

	// Validate authorized flag for dangerous commands
	if isDangerousCommand(cmd) && !authorized {
		return fmt.Errorf("offensive security operations require --authorized flag. " +
			"Only use this tool on systems you own or have explicit written permission to test. " +
			"Unauthorized access to computer systems is illegal.")
	}

	logger.Debugf("%s initialized - version: %s", constants.AppName, constants.Version)
	return nil
}

func isDangerousCommand(cmd *cobra.Command) bool {
	dangerous := []string{"poc", "brute", "exploit"}
	name := cmd.Name()
	for _, d := range dangerous {
		if name == d {
			return true
		}
	}
	// Check parent commands
	if cmd.HasParent() {
		return isDangerousCommand(cmd.Parent())
	}
	return false
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s version %s\n", constants.AppName, constants.Version)
			fmt.Printf("  Build Time: %s\n", constants.BuildTime)
			fmt.Printf("  Git Commit: %s\n", constants.GitCommit)
			fmt.Printf("  Git Branch: %s\n", constants.GitBranch)
			fmt.Printf("  Go Version: %s\n", constants.GoVersion)
		},
	}
}
