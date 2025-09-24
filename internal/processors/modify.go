package processors

import (
	"fmt"

	"github.com/pterm/pterm"

	"github.com/callmegreg/gh-security-config/internal/api"
	"github.com/callmegreg/gh-security-config/internal/types"
)

// ModifyProcessor implements OrganizationProcessor for the modify command
type ModifyProcessor struct {
	ConfigName     string
	NewName        string
	NewDescription string
	NewSettings    map[string]interface{}
}

// ProcessOrganization processes a single organization for the modify command
func (mp *ModifyProcessor) ProcessOrganization(org string) types.ProcessingResult {
	// Check membership using the shared validation function
	if skipResult := api.ValidateMembershipAndSkip(org); skipResult != nil {
		return *skipResult
	}

	updated, err := mp.modifyConfigurationInOrg(org)
	if err != nil {
		return types.ProcessingResult{Organization: org, Error: err}
	}
	if !updated {
		// Configuration was not found, already logged as warning in modifyConfigurationInOrg
		return types.ProcessingResult{Organization: org, Skipped: true}
	}

	return types.ProcessingResult{Organization: org, Success: true}
}

// modifyConfigurationInOrg updates a configuration in an organization
func (mp *ModifyProcessor) modifyConfigurationInOrg(org string) (bool, error) {
	// First, fetch security configurations for the organization
	configs, err := api.FetchSecurityConfigurations(org)
	if err != nil {
		return false, fmt.Errorf("failed to fetch security configurations: %w", err)
	}

	// Find the configuration by name
	configID, found := api.FindConfigurationByName(configs, mp.ConfigName)
	if !found {
		pterm.Warning.Printf("Configuration '%s' not found in organization '%s', skipping\n", mp.ConfigName, org)
		return false, nil // Not an error, just skip this org
	}

	// Update the configuration
	err = api.UpdateSecurityConfiguration(org, configID, mp.NewName, mp.NewDescription, mp.NewSettings)
	if err != nil {
		return false, fmt.Errorf("failed to update security configuration: %w", err)
	}

	return true, nil
}
