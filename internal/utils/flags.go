package utils

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// GetCommonFlags extracts common flags used across all commands
type CommonFlags struct {
	OrgListPath         string
	Concurrency         int
	DependabotAvailable *bool
}

// ExtractCommonFlags gets org-list and concurrency flags from command
func ExtractCommonFlags(cmd *cobra.Command) (*CommonFlags, error) {
	orgListPath, err := cmd.Flags().GetString("org-list")
	if err != nil {
		return nil, err
	}

	concurrency, err := cmd.Flags().GetInt("concurrency")
	if err != nil {
		return nil, err
	}

	dependabotAvailableFlag, err := cmd.Flags().GetString("dependabot-available")
	if err != nil {
		return nil, err
	}

	var dependabotAvailable *bool
	if dependabotAvailableFlag != "" {
		if dependabotAvailableFlag == "true" {
			val := true
			dependabotAvailable = &val
		} else if dependabotAvailableFlag == "false" {
			val := false
			dependabotAvailable = &val
		} else {
			return nil, fmt.Errorf("invalid value for dependabot-available flag: %s (must be 'true' or 'false')", dependabotAvailableFlag)
		}
	}

	return &CommonFlags{
		OrgListPath:         orgListPath,
		Concurrency:         concurrency,
		DependabotAvailable: dependabotAvailable,
	}, nil
}

// ValidateCSVEarly validates CSV file if provided
func ValidateCSVEarly(orgListPath string) error {
	if orgListPath != "" {
		orgs, err := ReadOrganizationsFromCSV(orgListPath)
		if err != nil {
			return fmt.Errorf("CSV validation failed: %w", err)
		}
		if len(orgs) == 0 {
			return fmt.Errorf("CSV file contains no valid organizations")
		}
	}
	return nil
}

// PrintCompletionHeader prints the completion header with results
func PrintCompletionHeader(operation string, successCount, skippedCount, errorCount int) {
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgGreen)).WithTextStyle(pterm.NewStyle(pterm.FgBlack)).Printf("%s Complete! (Success: %d, Skipped: %d, Errors: %d)", operation, successCount, skippedCount, errorCount)
}
