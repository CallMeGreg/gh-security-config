package processors

import (
	"fmt"

	"github.com/pterm/pterm"

	"github.com/callmegreg/gh-security-config/internal/api"
	"github.com/callmegreg/gh-security-config/internal/types"
)

// ApplyProcessor implements OrganizationProcessor for the apply command
type ApplyProcessor struct {
	ConfigName        string
	ConfigDescription string
	Settings          map[string]interface{}
	Scope             string
	SetAsDefault      bool
}

// ProcessOrganization processes a single organization for the apply command
func (ap *ApplyProcessor) ProcessOrganization(org string) types.ProcessingResult {
	// Check membership using the shared validation function
	if skipResult := api.ValidateMembershipAndSkip(org); skipResult != nil {
		return *skipResult
	}

	result := ap.processOrganization(org)
	return result
}

// processOrganization handles the core organization processing logic
func (ap *ApplyProcessor) processOrganization(org string) types.ProcessingResult {
	// Check if a configuration with the same name already exists
	configs, err := api.FetchSecurityConfigurations(org)
	if err != nil {
		return types.ProcessingResult{Organization: org, Error: fmt.Errorf("failed to fetch existing security configurations: %w", err)}
	}

	// Check if configuration already exists
	existingConfigID, exists := api.FindConfigurationByName(configs, ap.ConfigName)

	if !exists {
		// Configuration doesn't exist, skip this organization
		pterm.Info.Printf("Configuration '%s' not found in organization '%s', skipping\n", ap.ConfigName, org)
		return types.ProcessingResult{Organization: org, Skipped: true}
	}

	if ap.Scope != "" {
		err = api.AttachConfigurationToRepos(org, existingConfigID, ap.Scope)
		if err != nil {
			return types.ProcessingResult{Organization: org, Error: fmt.Errorf("failed to attach configuration to repositories: %w", err)}
		}
	}

	// Set as default if requested
	if ap.SetAsDefault {
		err = api.SetConfigurationAsDefault(org, existingConfigID)
		if err != nil {
			return types.ProcessingResult{Organization: org, Error: fmt.Errorf("failed to set configuration as default: %w", err)}
		}
	}

	return types.ProcessingResult{Organization: org, Success: true}
}
