package processors

import (
	"fmt"

	"github.com/pterm/pterm"

	"github.com/callmegreg/gh-security-config/internal/api"
	"github.com/callmegreg/gh-security-config/internal/types"
)

// GenerateProcessor implements OrganizationProcessor for the generate command
type GenerateProcessor struct {
	ConfigName        string
	ConfigDescription string
	Settings          map[string]interface{}
	Scope             string
	SetAsDefault      bool
	Force             bool
}

// ProcessOrganization processes a single organization for the generate command
func (gp *GenerateProcessor) ProcessOrganization(org string) types.ProcessingResult {
	// Check membership using the shared validation function
	if skipResult := api.ValidateMembershipAndSkip(org); skipResult != nil {
		return *skipResult
	}

	err := gp.processOrganization(org)
	if err != nil {
		return types.ProcessingResult{Organization: org, Error: err}
	}

	return types.ProcessingResult{Organization: org, Success: true}
}

// processOrganization handles the core organization processing logic
func (gp *GenerateProcessor) processOrganization(org string) error {
	// Check if a configuration with the same name already exists
	configs, err := api.FetchSecurityConfigurations(org)
	if err != nil {
		return fmt.Errorf("failed to fetch existing security configurations: %w", err)
	}

	// Check if configuration already exists
	existingConfigID, exists := api.FindConfigurationByName(configs, gp.ConfigName)
	if exists {
		if gp.Force {
			// Delete the existing configuration
			pterm.Info.Printf("Force flag enabled: deleting existing configuration '%s' from organization '%s'\n", gp.ConfigName, org)
			err = api.DeleteSecurityConfiguration(org, existingConfigID)
			if err != nil {
				return fmt.Errorf("failed to delete existing security configuration: %w", err)
			}
		} else {
			return &types.ConfigurationExistsError{
				ConfigName: gp.ConfigName,
				OrgName:    org,
			}
		}
	}

	// Create security configuration
	configID, err := api.CreateSecurityConfiguration(org, gp.ConfigName, gp.ConfigDescription, gp.Settings)
	if err != nil {
		return fmt.Errorf("failed to create security configuration: %w", err)
	}

	// Attach configuration to repositories only if scope is not "none"
	if gp.Scope != "none" {
		err = api.AttachConfigurationToRepos(org, configID, gp.Scope)
		if err != nil {
			return fmt.Errorf("failed to attach configuration to repositories: %w", err)
		}
	}

	// Set as default if requested
	if gp.SetAsDefault {
		err = api.SetConfigurationAsDefault(org, configID)
		if err != nil {
			return fmt.Errorf("failed to set configuration as default: %w", err)
		}
	}

	return nil
}
