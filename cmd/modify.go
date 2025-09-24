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

var modifyCmd = &cobra.Command{
	Use:   "modify",
	Short: "Modify existing security configurations across enterprise organizations",
	Long:  "Interactive command to update existing security configurations across all organizations in an enterprise",
	RunE:  runModify,
}

func runModify(cmd *cobra.Command, args []string) error {
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgMagenta)).WithTextStyle(pterm.NewStyle(pterm.FgWhite)).Println("GitHub Enterprise Security Configuration Modification")
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

	// Validate concurrency
	if err := utils.ValidateConcurrency(commonFlags.Concurrency); err != nil {
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

	// Check Dependabot availability
	dependabotAvailable, err := ui.GetDependabotAvailability(commonFlags.DependabotAvailable)
	if err != nil {
		return err
	}

	// Fetch organizations (from CSV or enterprise API)
	orgs, err := api.GetOrganizations(enterprise, commonFlags.OrgListPath)
	if err != nil {
		return err
	}

	if len(orgs) == 0 {
		ui.ShowNoOrganizationsWarning(commonFlags.OrgListPath)
		return nil
	}

	// Get security configuration name to modify
	configName, err := ui.GetConfigNameForModification()
	if err != nil {
		return err
	}

	// Fetch existing configuration details from first accessible organization to show current settings
	var currentSettings map[string]interface{}
	var currentDescription string
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
			configDetails, err := api.GetSecurityConfigurationDetails(org, configID)
			if err == nil {
				currentSettings = configDetails.Settings
				currentDescription = configDetails.Description
				break
			}
		}
	}

	if currentSettings == nil {
		pterm.Warning.Printf("Configuration '%s' not found in any accessible organizations.\n", configName)
		return fmt.Errorf("configuration '%s' not found", configName)
	}

	// Show current settings and get new settings
	pterm.Info.Println("Current configuration settings:")
	ui.DisplayCurrentSettings(currentSettings, currentDescription)
	pterm.Println()

	// Get new name
	newName, err := ui.GetUpdatedName(configName)
	if err != nil {
		return err
	}

	// Get new description
	newDescription, err := ui.GetUpdatedDescription(currentDescription)
	if err != nil {
		return err
	}

	// Get updated security settings
	newSettings, err := ui.GetSecuritySettingsForUpdate(currentSettings, dependabotAvailable)
	if err != nil {
		return err
	}

	// Confirm before proceeding
	confirmed, err := ui.ConfirmModifyOperation(orgs, configName, newName, currentDescription, newDescription, currentSettings, newSettings)
	if err != nil {
		return err
	}

	if !confirmed {
		ui.ShowOperationCancelled()
		return nil
	}

	// Process each organization
	ui.ShowProcessingStart(len(orgs), commonFlags.Concurrency)

	// Create processor for modify command
	processor := &processors.ModifyProcessor{
		ConfigName:     configName,
		NewName:        newName,
		NewDescription: newDescription,
		NewSettings:    newSettings,
	}

	// Use concurrent processor
	concurrentProcessor := processors.NewConcurrentProcessor(orgs, processor, commonFlags.Concurrency)
	successCount, skippedCount, errorCount := concurrentProcessor.Process()

	utils.PrintCompletionHeader("Security Configuration Modification", successCount, skippedCount, errorCount)

	return nil
}
