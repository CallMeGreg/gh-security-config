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
		DisableDefaultCmd: true,
	},
}

func init() {
	// Add persistent flags that are common to all commands
	rootCmd.PersistentFlags().String("org-list", "", "Path to CSV file containing organization names to target (one per line, no header)")
	rootCmd.PersistentFlags().Int("concurrency", 1, "Number of concurrent requests (1-20)")

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
