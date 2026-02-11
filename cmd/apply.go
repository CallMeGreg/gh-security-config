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
	Long: `Interactive command to apply an existing security configuration to organizations in an enterprise.

For GHES 3.16+, this command supports both organization-level and enterprise-level security configurations.`,
	RunE: runApply,
}

func init() {
	// Add template-org flag specific to apply command
	applyCmd.Flags().StringP("template-org", "t", "", "Template organization to fetch security configurations from (required)")
}

func runApply(cmd *cobra.Command, args []string) error {
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).WithTextStyle(pterm.NewStyle(pterm.FgWhite)).Println("GitHub Enterprise Security Configuration Application")
	pterm.Println()

	// Extract common flags
	commonFlags, err := utils.ExtractCommonFlags(cmd)
	if err != nil {
		return err
	}

	// Validate org targeting flags (optional for apply command)
	if err := utils.ValidateOrgFlagsOptional(commonFlags); err != nil {
		return err
	}

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

	templateOrgFlag, err := cmd.Flags().GetString("template-org")
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

	// Get template organization name
	templateOrg, err := ui.GetTemplateOrgInput(templateOrgFlag)
	if err != nil {
		return err
	}

	pterm.Info.Printf("Using template organization: %s\n", templateOrg)

	// Get GHES version from /meta endpoint to determine if enterprise configurations are available
	pterm.Info.Println("Detecting GitHub Enterprise Server version...")
	ghesVersion, err := api.GetGHESVersion()
	if err != nil {
		pterm.Warning.Printf("Could not detect GHES version: %v\n", err)
		pterm.Info.Println("Assuming GitHub Enterprise Cloud (GHEC) or enterprise configurations not available")
		ghesVersion = ""
	} else if ghesVersion != "" {
		pterm.Success.Printf("Detected GHES version: %s\n", ghesVersion)
	} else {
		pterm.Info.Println("Detected GitHub Enterprise Cloud (GHEC)")
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

	// Collect available configurations from both enterprise and template organization
	var orgConfigNames []string
	var enterpriseConfigNames []string
	enterpriseConfigMap := make(map[string]types.SecurityConfiguration)

	// Fetch enterprise configurations if GHES 3.16+
	if api.SupportsEnterpriseConfigurations(ghesVersion) {
		pterm.Info.Println("Fetching enterprise security configurations...")
		enterpriseConfigs, err := api.FetchEnterpriseSecurityConfigurations(enterprise)
		if err != nil {
			pterm.Warning.Printf("Could not fetch enterprise configurations: %v\n", err)
		} else {
			for _, config := range enterpriseConfigs {
				enterpriseConfigNames = append(enterpriseConfigNames, config.Name)
				enterpriseConfigMap[config.Name] = config
			}
			if len(enterpriseConfigs) > 0 {
				pterm.Success.Printf("Found %d enterprise security configuration(s)\n", len(enterpriseConfigs))
			}
		}
	}

	// Fetch org-level configuration names from template organization only
	pterm.Info.Printf("Fetching security configurations from template organization '%s'...\n", templateOrg)
	status, err := api.CheckSingleOrganizationMembership(templateOrg)
	if err != nil || !status.IsMember || !status.IsOwner {
		if err != nil {
			pterm.Warning.Printf("Could not access template organization '%s': %v\n", templateOrg, err)
		} else {
			pterm.Warning.Printf("You must be an owner of template organization '%s' to fetch configurations\n", templateOrg)
		}
	} else {
		configs, err := api.FetchSecurityConfigurations(templateOrg)
		if err != nil {
			pterm.Warning.Printf("Could not fetch configurations from template organization '%s': %v\n", templateOrg, err)
		} else {
			for _, config := range configs {
				// Only add organization-level configs (not enterprise configs shown at org level)
				if config.TargetType != "enterprise" {
					orgConfigNames = append(orgConfigNames, config.Name)
				}
			}
			if len(orgConfigNames) > 0 {
				pterm.Success.Printf("Found %d organization security configuration(s) in template org\n", len(orgConfigNames))
			}
		}
	}

	// Let user select a configuration
	var configName string
	var targetType string
	if len(enterpriseConfigNames) > 0 || len(orgConfigNames) > 0 {
		configName, targetType, err = ui.SelectConfigurationFromList(orgConfigNames, enterpriseConfigNames)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no security configurations found at enterprise or organization level")
	}

	// Get configuration details based on target type
	var configDetails *types.SecurityConfigurationDetails
	var sourceOrg string

	if targetType == "enterprise" {
		// Get enterprise configuration details
		enterpriseConfig, exists := enterpriseConfigMap[configName]
		if !exists {
			return fmt.Errorf("enterprise configuration '%s' not found in cached configurations", configName)
		}
		configDetails, err = api.GetEnterpriseSecurityConfigurationDetails(enterprise, enterpriseConfig.ID)
		if err != nil {
			return fmt.Errorf("failed to get enterprise configuration details: %w", err)
		}
		pterm.Info.Printf("Selected enterprise configuration: '%s'\n", configName)
	} else {
		// Get organization configuration details from template org
		configs, err := api.FetchSecurityConfigurations(templateOrg)
		if err != nil {
			return fmt.Errorf("failed to fetch configurations from template org: %w", err)
		}

		configID, found := api.FindConfigurationByName(configs, configName)
		if !found {
			return fmt.Errorf("configuration '%s' not found in template organization '%s'", configName, templateOrg)
		}

		details, err := api.GetSecurityConfigurationDetails(templateOrg, configID)
		if err != nil {
			return fmt.Errorf("failed to get configuration details: %w", err)
		}
		configDetails = details
		sourceOrg = templateOrg
		pterm.Info.Printf("Selected organization configuration '%s' from template org '%s'\n", configName, sourceOrg)
	}

	// Show configuration details
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
		ConfigName:         configName,
		ConfigDescription:  configDetails.Description,
		Settings:           configDetails.Settings,
		Scope:              scope,
		SetAsDefault:       setAsDefault,
		IsEnterpriseConfig: targetType == "enterprise",
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

	// Build and display replication command
	replicationFlags := map[string]interface{}{
		"enterprise-slug":              enterprise,
		"github-enterprise-server-url": serverURL,
		"template-org":                 templateOrg,
		"concurrency":                  commonFlags.Concurrency,
		"delay":                        commonFlags.Delay,
	}

	// Add org targeting flags
	if commonFlags.Org != "" {
		replicationFlags["org"] = commonFlags.Org
	} else if commonFlags.OrgListPath != "" {
		replicationFlags["org-list"] = commonFlags.OrgListPath
	} else if commonFlags.AllOrgs {
		replicationFlags["all-orgs"] = true
	}

	replicationCommand := utils.BuildReplicationCommand("apply", replicationFlags)
	utils.ShowReplicationCommand(replicationCommand)

	return nil
}
