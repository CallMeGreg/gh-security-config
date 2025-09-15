package processors

import (
	"fmt"

	"github.com/pterm/pterm"

	"github.com/callmegreg/gh-security-config/internal/api"
	"github.com/callmegreg/gh-security-config/internal/types"
)

// DeleteProcessor implements OrganizationProcessor for the delete command
type DeleteProcessor struct {
	ConfigName string
}

// ProcessOrganization processes a single organization for the delete command
func (dp *DeleteProcessor) ProcessOrganization(org string) types.ProcessingResult {
	// Check membership using the shared validation function
	if skipResult := api.ValidateMembershipAndSkip(org); skipResult != nil {
		return *skipResult
	}

	deleted, err := dp.deleteConfigurationFromOrg(org)
	if err != nil {
		return types.ProcessingResult{Organization: org, Error: err}
	}
	if !deleted {
		// Configuration was not found, already logged as warning in deleteConfigurationFromOrg
		return types.ProcessingResult{Organization: org, Skipped: true}
	}

	return types.ProcessingResult{Organization: org, Success: true}
}

// deleteConfigurationFromOrg deletes a configuration from an organization
func (dp *DeleteProcessor) deleteConfigurationFromOrg(org string) (bool, error) {
	// First, fetch security configurations for the organization
	configs, err := api.FetchSecurityConfigurations(org)
	if err != nil {
		return false, fmt.Errorf("failed to fetch security configurations: %w", err)
	}

	// Find the configuration by name
	configID, found := api.FindConfigurationByName(configs, dp.ConfigName)
	if !found {
		pterm.Warning.Printf("Configuration '%s' not found in organization '%s', skipping\n", dp.ConfigName, org)
		return false, nil // Not an error, just skip this org
	}

	// Delete the configuration
	err = api.DeleteSecurityConfiguration(org, configID)
	if err != nil {
		return false, fmt.Errorf("failed to delete security configuration: %w", err)
	}

	return true, nil
}
