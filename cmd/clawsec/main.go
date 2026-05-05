package main

import (
	"fmt"
	"os"

	"github.com/clawsec/clawsec/internal/cli"

	// Register product adapters
	_ "github.com/clawsec/clawsec/pkg/products/cloudwalker"
	_ "github.com/clawsec/clawsec/pkg/products/ddr"
	_ "github.com/clawsec/clawsec/pkg/products/safeline"
	_ "github.com/clawsec/clawsec/pkg/products/tanswer"
	_ "github.com/clawsec/clawsec/pkg/products/xray"
)

func main() {
	rootCmd := cli.NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "[%s] Error: %v\n", "ClawSec", err)
		os.Exit(1)
	}
}
