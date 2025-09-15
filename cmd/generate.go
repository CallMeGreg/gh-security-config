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
	Long:  "Interactive command to create security configurations and apply them to organizations in an enterprise. Optionally copy an existing configuration from another organization using --copy-from-org.",
	RunE:  runGenerate,
}

func init() {
	// Command-specific flags
	generateCmd.Flags().Bool("force", false, "Force deletion of existing configurations with the same name before creating new ones")
	generateCmd.Flags().String("copy-from-org", "", "Organization name to copy an existing configuration from")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgBlue)).WithTextStyle(pterm.NewStyle(pterm.FgWhite)).Println("GitHub Enterprise Security Configuration Generator")
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

	// Get generate-specific flags
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	copyFromOrg, err := cmd.Flags().GetString("copy-from-org")
	if err != nil {
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

		settings, err = ui.GetSecuritySettings()
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

	// Process each organization
	ui.ShowProcessingStart(len(orgs), commonFlags.Concurrency)

	// Create processor for generate command
	processor := &processors.GenerateProcessor{
		ConfigName:        configName,
		ConfigDescription: configDescription,
		Settings:          settings,
		Scope:             scope,
		SetAsDefault:      setAsDefault,
		Force:             force,
	}

	// Use concurrent processor
	concurrentProcessor := processors.NewConcurrentProcessor(orgs, processor, commonFlags.Concurrency)
	successCount, skippedCount, errorCount := concurrentProcessor.Process()

	utils.PrintCompletionHeader("Security Configuration Generation", successCount, skippedCount, errorCount)

	return nil
}
