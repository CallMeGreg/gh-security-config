package cmd

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/callmegreg/gh-security-config/internal/api"
	"github.com/callmegreg/gh-security-config/internal/processors"
	"github.com/callmegreg/gh-security-config/internal/ui"
	"github.com/callmegreg/gh-security-config/internal/utils"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete security configurations across enterprise organizations",
	Long:  "Interactive command to delete security configurations from all organizations in an enterprise",
	RunE:  runDelete,
}

func runDelete(cmd *cobra.Command, args []string) error {
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgRed)).WithTextStyle(pterm.NewStyle(pterm.FgWhite)).Println("GitHub Enterprise Security Configuration Deletion")
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

	// Get enterprise name
	enterprise, err := ui.GetEnterpriseInput()
	if err != nil {
		return err
	}

	// Get GitHub Enterprise Server URL if needed
	serverURL, err := ui.GetServerURLInput()
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

	// Get security configuration name to delete
	configName, err := ui.GetConfigNameForDeletion()
	if err != nil {
		return err
	}

	// Confirm before proceeding
	confirmed, err := ui.ConfirmDeleteOperation(orgs, configName)
	if err != nil {
		return err
	}

	if !confirmed {
		ui.ShowOperationCancelled()
		return nil
	}

	// Process each organization
	ui.ShowProcessingStart(len(orgs), commonFlags.Concurrency)

	// Create processor for delete command
	processor := &processors.DeleteProcessor{
		ConfigName: configName,
	}

	// Use concurrent processor
	concurrentProcessor := processors.NewConcurrentProcessor(orgs, processor, commonFlags.Concurrency)
	successCount, skippedCount, errorCount := concurrentProcessor.Process()

	utils.PrintCompletionHeader("Security Configuration Deletion", successCount, skippedCount, errorCount)

	return nil
}
