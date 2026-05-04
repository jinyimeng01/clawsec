package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newBruteCommand() *cobra.Command {
	var (
		targets      []string
		usernames    []string
		passwords    []string
		userFile     string
		passFile     string
		threads      int
		timeout      int
		delay        int
		stopOnSuccess bool
	)

	bruteCmd := &cobra.Command{
		Use:   "brute",
		Short: "Password brute-forcing engine (10+ protocols)",
		Long: `High-performance password brute-forcing engine supporting multiple protocols.

Supported Protocols:
  ssh      - SSH password/key authentication
  ftp      - FTP authentication
  rdp      - RDP authentication
  mysql    - MySQL/MariaDB authentication
  redis    - Redis authentication
  mongodb  - MongoDB authentication
  postgres - PostgreSQL authentication
  mssql    - Microsoft SQL Server authentication
  smb      - SMB/CIFS authentication
  ldap     - LDAP authentication
  http     - HTTP Basic/Digest authentication

Examples:
  # SSH brute force with single credential
  clawsec brute ssh -t 10.0.0.1 -u root -P passwords.txt

  # Brute force multiple targets with multiple users
  clawsec brute ssh -t targets.txt -U users.txt -P passwords.txt --threads 100

  # Redis brute force
  clawsec brute redis -t 10.0.0.1 -P passwords.txt

  # HTTP Basic auth brute force
  clawsec brute http -t http://target.com -u admin -P passwords.txt`,
	}

	protocols := []string{"ssh", "ftp", "rdp", "mysql", "redis", "mongodb", "postgres", "mssql", "smb", "ldap", "http"}

	for _, proto := range protocols {
		proto := proto // capture range variable
		cmd := &cobra.Command{
			Use:   proto,
			Short: fmt.Sprintf("%s password brute-forcing", strings.ToUpper(proto)),
			RunE: func(cmd *cobra.Command, args []string) error {
				if len(targets) == 0 {
					return fmt.Errorf("no targets specified, use -t flag")
				}
				if len(usernames) == 0 && userFile == "" {
					return fmt.Errorf("no usernames specified, use -u or -U flag")
				}
				if len(passwords) == 0 && passFile == "" {
					return fmt.Errorf("no passwords specified, use -p or -P flag")
				}

				fmt.Printf("[INF] Starting %s brute force attack\n", strings.ToUpper(proto))
				fmt.Printf("[INF] Targets: %d | Threads: %d | Timeout: %ds\n",
					len(targets), threads, timeout)
				fmt.Printf("[INF] %s brute force engine - implementation in progress (Phase 4)\n",
					strings.ToUpper(proto))
				return nil
			},
		}

		cmd.Flags().StringArrayVarP(&targets, "target", "t", nil, fmt.Sprintf("target %s hosts", proto))
		cmd.Flags().StringArrayVarP(&usernames, "username", "u", nil, "usernames to try")
		cmd.Flags().StringArrayVarP(&passwords, "password", "p", nil, "passwords to try")
		cmd.Flags().StringVarP(&userFile, "user-file", "U", "", "username dictionary file")
		cmd.Flags().StringVarP(&passFile, "pass-file", "P", "", "password dictionary file")
		cmd.Flags().IntVar(&threads, "threads", 10, "concurrent threads")
		cmd.Flags().IntVar(&timeout, "timeout", 5, "connection timeout in seconds")
		cmd.Flags().IntVar(&delay, "delay", 0, "delay between attempts in milliseconds")
		cmd.Flags().BoolVar(&stopOnSuccess, "stop-on-success", false, "stop after first successful login")
		cmd.MarkFlagRequired("target")

		bruteCmd.AddCommand(cmd)
	}

	return bruteCmd
}
