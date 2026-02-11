package cmd

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/callmegreg/gh-security-config/internal/api"
	"github.com/callmegreg/gh-security-config/internal/processors"
	"github.com/callmegreg/gh-security-config/internal/ui"
	"github.com/callmegreg/gh-security-config/internal/utils"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate and apply security configurations across enterprise organizations",
	Long:  "Interactive command to create security configurations and apply them to organizations in an enterprise.",
	RunE:  runGenerate,
}

func init() {
	// Command-specific flags
	generateCmd.Flags().BoolP("force", "f", false, "Force deletion of existing configurations with the same name before creating new ones")
	generateCmd.Flags().StringP("copy-from-org", "o", "", "Organization name to copy an existing configuration from")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgBlue)).WithTextStyle(pterm.NewStyle(pterm.FgWhite)).Println("GitHub Enterprise Security Configuration Generator")
	pterm.Println()

	// Extract common flags
	commonFlags, err := utils.ExtractCommonFlags(cmd)
	if err != nil {
		return err
	}

	// Validate org targeting flags (optional for generate command)
	if err := utils.ValidateOrgFlagsOptional(commonFlags); err != nil {
		return err
	}

	// Get generate-specific flags
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	copyFromOrg, err := cmd.Flags().GetString("copy-from-org")
	if err != nil {
		return err
	}

	// Validate concurrency and delay flags
	if err := utils.ValidateConcurrency(commonFlags.Concurrency); err != nil {
		return err
	}
	if err := utils.ValidateDelay(commonFlags.Delay); err != nil {
		return err
	}
	if err := utils.ValidateConcurrencyAndDelay(commonFlags.Concurrency, commonFlags.Delay); err != nil {
		return err
	}

	// Get flag values for enterprise settings
	enterpriseFlag, err := cmd.Flags().GetString("enterprise-slug")
	if err != nil {
		return err
	}

	serverURLFlag, err := cmd.Flags().GetString("github-enterprise-server-url")
	if err != nil {
		return err
	}

	// Get enterprise name
	enterprise, err := ui.GetEnterpriseInput(enterpriseFlag)
	if err != nil {
		return err
	}

	// Get GitHub Enterprise Server URL if needed
	serverURL, err := ui.GetServerURLInput(serverURLFlag)
	if err != nil {
		return err
	}

	// Set hostname if using GitHub Enterprise Server
	ui.SetupGitHubHost(serverURL)

	// If no org targeting method is provided, prompt user to select one
	if !utils.HasOrgTargeting(commonFlags) {
		targetingMethod, err := ui.SelectOrgTargetingMethod()
		if err != nil {
			return err
		}

		switch targetingMethod {
		case "all-orgs":
			commonFlags.AllOrgs = true
		case "single-org":
			orgName, err := ui.GetSingleOrgName()
			if err != nil {
				return err
			}
			commonFlags.Org = orgName
		case "org-list":
			csvPath, err := ui.GetOrgListPath()
			if err != nil {
				return err
			}
			commonFlags.OrgListPath = csvPath
			// Validate the CSV file
			if err := utils.ValidateOrgFlagsOptional(commonFlags); err != nil {
				return err
			}
		}
	}

	// Check Dependabot availability
	dependabotAlertsAvailable, err := ui.GetDependabotAlertsAvailability(commonFlags.DependabotAlertsAvailable)
	if err != nil {
		return err
	}

	dependabotSecurityUpdatesAvailable, err := ui.GetDependabotSecurityUpdatesAvailability(commonFlags.DependabotSecurityUpdatesAvailable)
	if err != nil {
		return err
	}

	// Fetch organizations
	orgs, err := api.GetOrganizations(enterprise, commonFlags.Org, commonFlags.OrgListPath, commonFlags.AllOrgs)
	if err != nil {
		return err
	}

	if len(orgs) == 0 {
		ui.ShowNoOrganizationsWarning(commonFlags)
		return nil
	}

	var configName, configDescription string
	var settings map[string]interface{}
	var scope string
	var setAsDefault bool

	// Check if we should copy from an existing organization
	if copyFromOrg != "" {
		// Filter out the source organization from target organizations to avoid copying to itself
		var filteredOrgs []string
		for _, org := range orgs {
			if org != copyFromOrg {
				filteredOrgs = append(filteredOrgs, org)
			}
		}

		if len(filteredOrgs) == 0 {
			return fmt.Errorf("no target organizations available after excluding source organization '%s'", copyFromOrg)
		}

		if len(filteredOrgs) < len(orgs) {
			pterm.Info.Printf("Excluding source organization '%s' from targets. Will process %d organizations.\n", copyFromOrg, len(filteredOrgs))
			orgs = filteredOrgs
		}

		// Copy configuration logic
		configName, configDescription, settings, scope, setAsDefault, err = ui.HandleCopyFromOrg(copyFromOrg)
		if err != nil {
			return err
		}
	} else {
		// Original logic for creating new configuration
		configName, configDescription, err = ui.GetSecurityConfigInput()
		if err != nil {
			return err
		}

		settings, err = ui.GetSecuritySettings(dependabotAlertsAvailable, dependabotSecurityUpdatesAvailable)
		if err != nil {
			return err
		}

		scope, err = ui.GetAttachmentScope()
		if err != nil {
			return err
		}

		setAsDefault, err = ui.GetDefaultSetting()
		if err != nil {
			return err
		}
	}

	// Confirm before proceeding
	confirmed, err := ui.ConfirmOperation(orgs, configName, configDescription, settings, scope, setAsDefault)
	if err != nil {
		return err
	}

	if !confirmed {
		ui.ShowOperationCancelled()
		return nil
	}

	// Create processor for generate command
	processor := &processors.GenerateProcessor{
		ConfigName:        configName,
		ConfigDescription: configDescription,
		Settings:          settings,
		Scope:             scope,
		SetAsDefault:      setAsDefault,
		Force:             force,
	}

	// Process each organization - use sequential processor if delay is specified
	var successCount, skippedCount, errorCount int
	if commonFlags.Delay > 0 {
		ui.ShowProcessingStartWithDelay(len(orgs), commonFlags.Delay)
		sequentialProcessor := processors.NewSequentialProcessor(orgs, processor, commonFlags.Delay)
		successCount, skippedCount, errorCount = sequentialProcessor.Process()
	} else {
		ui.ShowProcessingStart(len(orgs), commonFlags.Concurrency)
		concurrentProcessor := processors.NewConcurrentProcessor(orgs, processor, commonFlags.Concurrency)
		successCount, skippedCount, errorCount = concurrentProcessor.Process()
	}

	utils.PrintCompletionHeader("Security Configuration Generation", successCount, skippedCount, errorCount)

	// Build and display replication command
	replicationFlags := map[string]interface{}{
		"enterprise-slug":                       enterprise,
		"github-enterprise-server-url":          serverURL,
		"dependabot-alerts-available":           fmt.Sprintf("%t", dependabotAlertsAvailable),
		"dependabot-security-updates-available": fmt.Sprintf("%t", dependabotSecurityUpdatesAvailable),
		"concurrency":                           commonFlags.Concurrency,
		"delay":                                 commonFlags.Delay,
		"force":                                 force,
	}

	// Add org targeting flags
	if commonFlags.Org != "" {
		replicationFlags["org"] = commonFlags.Org
	} else if commonFlags.OrgListPath != "" {
		replicationFlags["org-list"] = commonFlags.OrgListPath
	} else if commonFlags.AllOrgs {
		replicationFlags["all-orgs"] = true
	}

	// Add copy-from-org flag if used
	if copyFromOrg != "" {
		replicationFlags["copy-from-org"] = copyFromOrg
	}

	replicationCommand := utils.BuildReplicationCommand("generate", replicationFlags)
	utils.ShowReplicationCommand(replicationCommand)

	return nil
}
