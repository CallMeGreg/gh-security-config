package utils

import (
	"fmt"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// GetCommonFlags extracts common flags used across all commands
type CommonFlags struct {
	Org                                string
	OrgListPath                        string
	AllOrgs                            bool
	Concurrency                        int
	Delay                              int
	DependabotAlertsAvailable          *bool
	DependabotSecurityUpdatesAvailable *bool
}

// ExtractCommonFlags gets org targeting, concurrency, and delay flags from command
func ExtractCommonFlags(cmd *cobra.Command) (*CommonFlags, error) {
	org, err := cmd.Flags().GetString("org")
	if err != nil {
		return nil, err
	}

	orgListPath, err := cmd.Flags().GetString("org-list")
	if err != nil {
		return nil, err
	}

	allOrgs, err := cmd.Flags().GetBool("all-orgs")
	if err != nil {
		return nil, err
	}

	concurrency, err := cmd.Flags().GetInt("concurrency")
	if err != nil {
		return nil, err
	}

	delay, err := cmd.Flags().GetInt("delay")
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
		Org:                                org,
		OrgListPath:                        orgListPath,
		AllOrgs:                            allOrgs,
		Concurrency:                        concurrency,
		Delay:                              delay,
		DependabotAlertsAvailable:          dependabotAlertsAvailable,
		DependabotSecurityUpdatesAvailable: dependabotSecurityUpdatesAvailable,
	}, nil
}

// ValidateOrgFlags validates org targeting flags and CSV file if provided
func ValidateOrgFlags(flags *CommonFlags) error {
	// Ensure at least one org targeting option is provided
	if flags.Org == "" && flags.OrgListPath == "" && !flags.AllOrgs {
		return fmt.Errorf("one of --org, --org-list, or --all-orgs must be specified")
	}

	// Validate CSV file early if provided
	if flags.OrgListPath != "" {
		orgs, err := ReadOrganizationsFromCSV(flags.OrgListPath)
		if err != nil {
			return fmt.Errorf("CSV validation failed: %w", err)
		}
		if len(orgs) == 0 {
			return fmt.Errorf("CSV file contains no valid organizations")
		}
	}

	// Validate single org name format
	if flags.Org != "" {
		if strings.Contains(flags.Org, " ") || strings.Contains(flags.Org, "/") {
			return fmt.Errorf("invalid organization name format: %s", flags.Org)
		}
	}

	return nil
}

// PrintCompletionHeader prints the completion header with results
func PrintCompletionHeader(operation string, successCount, skippedCount, errorCount int) {
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgGreen)).WithTextStyle(pterm.NewStyle(pterm.FgBlack)).Printf("%s Complete! (Success: %d, Skipped: %d, Errors: %d)", operation, successCount, skippedCount, errorCount)
}
