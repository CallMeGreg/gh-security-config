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

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete security configurations across enterprise organizations",
	Long:  "Interactive command to delete security configurations from organizations in an enterprise",
	RunE:  runDelete,
}

func init() {
	// Add template-org flag specific to delete command
	deleteCmd.Flags().StringP("template-org", "t", "", "Template organization to fetch security configurations from (required)")
}

func runDelete(cmd *cobra.Command, args []string) error {
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgRed)).WithTextStyle(pterm.NewStyle(pterm.FgWhite)).Println("GitHub Security Configuration Deleter")
	pterm.Println()

	// Extract common flags
	commonFlags, err := utils.ExtractCommonFlags(cmd)
	if err != nil {
		return err
	}

	// Validate org targeting flags (optional for delete command)
	if err := utils.ValidateOrgFlagsOptional(commonFlags); err != nil {
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

	templateOrgFlag, err := cmd.Flags().GetString("template-org")
	if err != nil {
		return err
	}

	configNameFlag, err := cmd.Flags().GetString("config-name")
	if err != nil {
		return err
	}

	force, err := extractSkipConfirmationFlag(cmd)
	if err != nil {
		return err
	}

	// Get enterprise name
	enterprise, err := ui.GetEnterpriseInput(enterpriseFlag)
	if err != nil {
		return err
	}

	// Get GitHub Enterprise URL if needed
	serverURL, err := ui.GetServerURLInput(serverURLFlag)
	if err != nil {
		return err
	}

	// Set hostname if using GitHub Enterprise Server
	ui.SetupGitHubHost(serverURL)

	// Collect available configurations from template organization
	var orgConfigNames []string

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

	// Get template organization name
	templateOrg, err := ui.GetTemplateOrgInput(templateOrgFlag)
	if err != nil {
		return err
	}

	pterm.Info.Printf("Using template organization: %s\n", templateOrg)

	// Fetch org-level configuration names from template organization only
	pterm.Info.Printf("Fetching security configurations from template organization '%s'...\n", templateOrg)
	status, err := api.CheckSingleOrganizationMembership(templateOrg)
	if err != nil {
		pterm.Warning.Printf("Could not access template organization '%s': %v\n", templateOrg, err)
	} else if !status.IsMember {
		pterm.Warning.Printf("You must be a member of template organization '%s' to fetch configurations\n", templateOrg)
	} else if !status.IsOwner {
		pterm.Warning.Printf("You must be an owner of template organization '%s' to fetch configurations\n", templateOrg)
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

	// Let user select a configuration to delete
	var configName string
	if len(orgConfigNames) > 0 {
		configName, err = ui.SelectConfigurationForDeletion(orgConfigNames, configNameFlag)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("no security configurations found in template organization '%s'", templateOrg)
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

	// Confirm before proceeding
	confirmed, err := ui.ConfirmDeleteOperation(orgs, configName, force)
	if err != nil {
		return err
	}

	if !confirmed {
		ui.ShowOperationCancelled()
		return nil
	}

	// Create processor for delete command
	processor := &processors.DeleteProcessor{
		ConfigName: configName,
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

	utils.PrintCompletionHeader("Security Configuration Deletion", successCount, skippedCount, errorCount)

	// Build and display replication command
	replicationFlags := map[string]interface{}{
		"enterprise-slug":              enterprise,
		"github-enterprise-server-url": serverURL,
		"template-org":                 templateOrg,
		"concurrency":                  commonFlags.Concurrency,
		"delay":                        commonFlags.Delay,
		"config-name":                  configName,
		"skip-confirmation-message":    fmt.Sprintf("%t", force),
	}

	// Add org targeting flags
	if commonFlags.Org != "" {
		replicationFlags["org"] = commonFlags.Org
	} else if commonFlags.OrgListPath != "" {
		replicationFlags["org-list"] = commonFlags.OrgListPath
	} else if commonFlags.AllOrgs {
		replicationFlags["all-orgs"] = true
	}

	replicationCommand := utils.BuildReplicationCommand("delete", replicationFlags)
	utils.ShowReplicationCommand(replicationCommand)

	return nil
}
