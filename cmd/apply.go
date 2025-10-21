package cmd

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/callmegreg/gh-security-config/internal/api"
	"github.com/callmegreg/gh-security-config/internal/processors"
	"github.com/callmegreg/gh-security-config/internal/types"
	"github.com/callmegreg/gh-security-config/internal/ui"
	"github.com/callmegreg/gh-security-config/internal/utils"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply existing security configurations to repositories",
	Long:  "Interactive command to apply an existing security configuration to specific repositories across organizations in an enterprise",
	RunE:  runApply,
}

func runApply(cmd *cobra.Command, args []string) error {
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).WithTextStyle(pterm.NewStyle(pterm.FgWhite)).Println("GitHub Enterprise Security Configuration Application")
	pterm.Println()

	// Extract common flags
	commonFlags, err := utils.ExtractCommonFlags(cmd)
	if err != nil {
		return err
	}

	// Validate CSV file early if provided
	if err := utils.ValidateCSVEarly(commonFlags.OrgListPath); err != nil {
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

	// Fetch organizations (from CSV or enterprise API)
	orgs, err := api.GetOrganizations(enterprise, commonFlags.OrgListPath)
	if err != nil {
		return err
	}

	if len(orgs) == 0 {
		ui.ShowNoOrganizationsWarning(commonFlags.OrgListPath)
		return nil
	}

	// Get security configuration name to apply
	configName, err := ui.GetConfigNameForApplication()
	if err != nil {
		return err
	}

	// Verify configuration exists in at least one organization and get its details
	var configDetails *types.SecurityConfigurationDetails
	var sourceOrg string
	for _, org := range orgs {
		// Check membership for this specific organization
		status, err := api.CheckSingleOrganizationMembership(org)
		if err != nil || !status.IsMember || !status.IsOwner {
			continue
		}

		configs, err := api.FetchSecurityConfigurations(org)
		if err != nil {
			continue
		}

		configID, found := api.FindConfigurationByName(configs, configName)
		if found {
			// Get detailed configuration
			details, err := api.GetSecurityConfigurationDetails(org, configID)
			if err == nil {
				configDetails = details
				sourceOrg = org
				break
			}
		}
	}

	if configDetails == nil {
		pterm.Warning.Printf("Configuration '%s' not found in any accessible organizations.\n", configName)
		return fmt.Errorf("configuration '%s' not found", configName)
	}

	// Show configuration details
	pterm.Info.Printf("Found configuration '%s' in organization '%s'\n", configName, sourceOrg)
	ui.DisplayCurrentSettings(configDetails.Settings, configDetails.Description)
	pterm.Println()

	// Get repository attachment scope (without 'none' option)
	scope, err := ui.GetAttachmentScopeForApplication()
	if err != nil {
		return err
	}

	// Get default setting
	setAsDefault, err := ui.GetDefaultSetting()
	if err != nil {
		return err
	}

	// Confirm before proceeding
	confirmed, err := ui.ConfirmApplyOperation(orgs, configName, configDetails.Description, configDetails.Settings, scope, setAsDefault)
	if err != nil {
		return err
	}

	if !confirmed {
		ui.ShowOperationCancelled()
		return nil
	}

	// Create processor for apply command
	processor := &processors.ApplyProcessor{
		ConfigName:        configName,
		ConfigDescription: configDetails.Description,
		Settings:          configDetails.Settings,
		Scope:             scope,
		SetAsDefault:      setAsDefault,
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

	utils.PrintCompletionHeader("Security Configuration Application", successCount, skippedCount, errorCount)

	return nil
}
