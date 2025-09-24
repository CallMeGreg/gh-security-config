package utils

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// GetCommonFlags extracts common flags used across all commands
type CommonFlags struct {
	OrgListPath                        string
	Concurrency                        int
	DependabotAlertsAvailable          *bool
	DependabotSecurityUpdatesAvailable *bool
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

	dependabotAlertsAvailableFlag, err := cmd.Flags().GetString("dependabot-alerts-available")
	if err != nil {
		return nil, err
	}

	dependabotSecurityUpdatesAvailableFlag, err := cmd.Flags().GetString("dependabot-security-updates-available")
	if err != nil {
		return nil, err
	}

	var dependabotAlertsAvailable *bool
	if dependabotAlertsAvailableFlag != "" {
		if dependabotAlertsAvailableFlag == "true" {
			val := true
			dependabotAlertsAvailable = &val
		} else if dependabotAlertsAvailableFlag == "false" {
			val := false
			dependabotAlertsAvailable = &val
		} else {
			return nil, fmt.Errorf("invalid value for dependabot-alerts-available flag: %s (must be 'true' or 'false')", dependabotAlertsAvailableFlag)
		}
	}

	var dependabotSecurityUpdatesAvailable *bool
	if dependabotSecurityUpdatesAvailableFlag != "" {
		if dependabotSecurityUpdatesAvailableFlag == "true" {
			val := true
			dependabotSecurityUpdatesAvailable = &val
		} else if dependabotSecurityUpdatesAvailableFlag == "false" {
			val := false
			dependabotSecurityUpdatesAvailable = &val
		} else {
			return nil, fmt.Errorf("invalid value for dependabot-security-updates-available flag: %s (must be 'true' or 'false')", dependabotSecurityUpdatesAvailableFlag)
		}
	}

	return &CommonFlags{
		OrgListPath:                        orgListPath,
		Concurrency:                        concurrency,
		DependabotAlertsAvailable:          dependabotAlertsAvailable,
		DependabotSecurityUpdatesAvailable: dependabotSecurityUpdatesAvailable,
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
