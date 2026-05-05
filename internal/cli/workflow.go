package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/clawsec/clawsec/internal/logger"
	"github.com/spf13/cobra"
)

func newWorkflowCommand() *cobra.Command {
	var (
		target    string
		objective string
		strategy  string
	)

	workflowCmd := &cobra.Command{
		Use:   "workflow",
		Short: "Automated penetration testing workflows",
		Long: `AI-driven automated penetration testing workflows.

The workflow engine uses AI to plan and execute multi-step attack chains,
integrating port scanning, PoC execution, and vulnerability verification.

Examples:
  # Full reconnaissance workflow
  clawsec workflow run -t 10.0.0.1 --objective "find all vulnerabilities"

  # Targeted exploit chain
  clawsec workflow run -t http://target.com --objective "achieve RCE"

  # Stealth assessment
  clawsec workflow run -t 10.0.0.0/24 --strategy stealth`,
	}

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run an automated penetration testing workflow",
		RunE: func(cmd *cobra.Command, args []string) error {
			if target == "" {
				return fmt.Errorf("target is required")
			}

			agent, err := initAIAgent()
			if err != nil {
				return fmt.Errorf("AI Agent required for workflow execution: %w", err)
			}

			logger.Infof("╔════════════════════════════════════════════════════════════╗")
			logger.Infof("║     ClawSec Automated Penetration Testing Workflow         ║")
			logger.Infof("╚════════════════════════════════════════════════════════════╝")
			logger.Infof("Target:    %s", target)
			logger.Infof("Objective: %s", objective)
			logger.Infof("Strategy:  %s", strategy)
			fmt.Println()

			// Phase 1: Reconnaissance
			logger.Infof("[Phase 1/4] Reconnaissance - Analyzing target...")
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			analyzeResult, err := agent.Call(ctx, "analyze", map[string]interface{}{
				"target": target,
			})
			cancel()
			if err != nil {
				logger.Warnf("AI analysis failed: %v", err)
			} else {
				resultMap, _ := analyzeResult.Result.(map[string]interface{})
				if risk, ok := resultMap["risk_score"].(float64); ok {
					logger.Infof("Risk Score: %.1f/10", risk)
				}
				if steps, ok := resultMap["recommended_next_steps"].([]interface{}); ok && len(steps) > 0 {
					logger.Infof("Recommended next steps:")
					for _, step := range steps {
						logger.Infof("  - %v", step)
					}
				}
			}

			// Phase 2: Port Scan
			logger.Infof("[Phase 2/4] Port Scanning - Identifying open services...")
			logger.Infof("Execute: clawsec scan port -t %s -p top1000", target)
			logger.Infof("(Run the above command manually or integrate with --ai flag)")

			// Phase 3: Vulnerability Verification
			logger.Infof("[Phase 3/4] Vulnerability Verification - Running PoC templates...")
			logger.Infof("Execute: clawsec poc run -d templates/poc -u %s --authorized", target)

			// Phase 4: Exploit Chain Building
			logger.Infof("[Phase 4/4] Exploit Chain - Building attack path...")
			ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
			chainResult, err := agent.Call(ctx, "chain", map[string]interface{}{
				"target":    target,
				"findings":  []map[string]string{},
				"objective": objective,
			})
			cancel()
			if err != nil {
				logger.Warnf("Chain building failed: %v", err)
			} else {
				resultMap, _ := chainResult.Result.(map[string]interface{})
				if steps, ok := resultMap["steps"].([]interface{}); ok {
					logger.Infof("Suggested exploit chain:")
					for _, step := range steps {
						if s, ok := step.(map[string]interface{}); ok {
							logger.Infof("  Step %v: %v (tool: %v)", s["step"], s["action"], s["tool"])
						}
					}
				}
			}

			fmt.Println()
			logger.Infof("Workflow planning complete.")
			logger.Infof("To execute the full workflow with real scanning:")
			logger.Infof("  clawsec scan port -t %s -p top1000 --ai", target)
			logger.Infof("  clawsec poc run -d templates/poc -u %s --ai --authorized", target)

			return nil
		},
	}

	runCmd.Flags().StringVarP(&target, "target", "t", "", "target to assess (required)")
	runCmd.Flags().StringVar(&objective, "objective", "find vulnerabilities", "penetration testing objective")
	runCmd.Flags().StringVar(&strategy, "strategy", "aggressive", "strategy: aggressive|stealth|comprehensive")
	runCmd.MarkFlagRequired("target")

	workflowCmd.AddCommand(runCmd)
	return workflowCmd
}
