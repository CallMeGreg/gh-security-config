package cmd

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "security-config",
	Short: "GitHub Security Configuration Management for Enterprises",
	Long:  "A GitHub CLI extension to manage security configurations across all organizations in an enterprise",
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
}

func init() {
	// Add persistent flags that are common to all commands
	// Organization targeting: three mutually exclusive options
	rootCmd.PersistentFlags().String("org", "", "Target a single organization by name")
	rootCmd.PersistentFlags().StringP("org-list", "l", "", "Path to CSV file containing organization names to target (one per line, no header)")
	rootCmd.PersistentFlags().Bool("all-orgs", false, "Target all organizations in the enterprise")

	rootCmd.PersistentFlags().IntP("concurrency", "c", 1, "Number of concurrent requests (1-20)")
	rootCmd.PersistentFlags().IntP("delay", "d", 0, "Delay in seconds between organizations (1-600, mutually exclusive with --concurrency)")
	rootCmd.PersistentFlags().StringP("enterprise-slug", "e", "", "GitHub Enterprise slug (e.g., github)")
	rootCmd.PersistentFlags().StringP("github-enterprise-server-url", "u", "", "GitHub Enterprise Server URL (e.g., github.company.com)")
	rootCmd.PersistentFlags().StringP("dependabot-alerts-available", "a", "", "Whether Dependabot Alerts are available in your GHES instance (true/false)")
	rootCmd.PersistentFlags().StringP("dependabot-security-updates-available", "s", "", "Whether Dependabot Security Updates are available in your GHES instance (true/false)")

	// Mark org targeting flags as mutually exclusive
	rootCmd.MarkFlagsMutuallyExclusive("org", "org-list", "all-orgs")

	// Mark concurrency and delay as mutually exclusive
	rootCmd.MarkFlagsMutuallyExclusive("concurrency", "delay")

	// Add subcommands
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(modifyCmd)
	rootCmd.AddCommand(applyCmd)
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		pterm.Error.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
