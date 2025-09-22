package ui

import (
	"fmt"

	"github.com/pterm/pterm"

	"github.com/callmegreg/gh-security-config/internal/api"
	"github.com/callmegreg/gh-security-config/internal/types"
)

// ConfirmOperation shows operation summary and asks for confirmation
func ConfirmOperation(orgs []string, configName, configDescription string, settings map[string]interface{}, scope string, setAsDefault bool) (bool, error) {
	pterm.Println()
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgYellow)).WithTextStyle(pterm.NewStyle(pterm.FgBlack)).Println("Operation Summary")

	pterm.Printf("Organizations: %d\n", len(orgs))
	pterm.Printf("Configuration Name: %s\n", pterm.Yellow(configName))
	pterm.Printf("Configuration Description: %s\n", pterm.Yellow(configDescription))
	pterm.Println()

	pterm.Info.Println("Security Settings:")
	for key, value := range settings {
		valueStr := fmt.Sprintf("%v", value)
		var coloredValue string

		switch valueStr {
		case "enabled", "enforced":
			coloredValue = pterm.Green(valueStr)
		case "disabled", "unenforced":
			coloredValue = pterm.Red(valueStr)
		case "not_set":
			coloredValue = pterm.Yellow(valueStr)
		default:
			coloredValue = pterm.Yellow(valueStr)
		}

		pterm.Printf("  %s: %s\n", pterm.Cyan(key), coloredValue)
	}
	pterm.Println()

	pterm.Printf("Attachment Scope: %s\n", pterm.Magenta(scope))
	pterm.Printf("Set as Default: %s\n", pterm.Cyan(fmt.Sprintf("%t", setAsDefault)))
	pterm.Println()

	confirmed, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Proceed with creating security configurations?").Show()
	if err != nil {
		return false, err
	}

	return confirmed, nil
}

// ConfirmDeleteOperation shows delete summary and asks for confirmation
func ConfirmDeleteOperation(orgs []string, configName string) (bool, error) {
	pterm.Println()
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgRed)).WithTextStyle(pterm.NewStyle(pterm.FgWhite)).Println("DELETE OPERATION SUMMARY")

	pterm.Printf("Organizations: %d\n", len(orgs))
	pterm.Printf("Configuration to Delete: %s\n", pterm.Red(configName))
	pterm.Println()

	pterm.Warning.Println("WARNING: This operation will delete the security configuration from ALL organizations in the enterprise.")
	pterm.Warning.Println("This action cannot be undone. Repositories will retain their settings but will no longer be associated with the configuration.")
	pterm.Println()

	confirmed, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Are you absolutely sure you want to proceed with deleting this configuration?").WithDefaultValue(false).Show()
	if err != nil {
		return false, err
	}

	return confirmed, nil
}

// ConfirmModifyOperation shows modify summary and asks for confirmation
func ConfirmModifyOperation(orgs []string, configName, currentDescription, newDescription string, currentSettings, newSettings map[string]interface{}) (bool, error) {
	pterm.Println()
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgYellow)).WithTextStyle(pterm.NewStyle(pterm.FgBlack)).Println("MODIFY OPERATION SUMMARY")

	pterm.Printf("Organizations: %d\n", len(orgs))
	pterm.Printf("Configuration to Modify: %s\n", pterm.Magenta(configName))
	pterm.Println()

	// Show changes
	pterm.Info.Println("Changes to be made:")

	// Description changes
	if currentDescription != newDescription {
		pterm.Printf("  Description: %s → %s\n", pterm.Red(currentDescription), pterm.Green(newDescription))
	} else {
		pterm.Printf("  Description: %s (no change)\n", pterm.Yellow(currentDescription))
	}

	// Setting changes
	for key, newValue := range newSettings {
		currentValue := fmt.Sprintf("%v", currentSettings[key])
		newValueStr := fmt.Sprintf("%v", newValue)

		if currentValue != newValueStr {
			pterm.Printf("  %s: %s → %s\n", pterm.Cyan(key), pterm.Red(currentValue), pterm.Green(newValueStr))
		} else {
			pterm.Printf("  %s: %s (no change)\n", pterm.Cyan(key), pterm.Yellow(currentValue))
		}
	}

	pterm.Println()

	confirmed, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Proceed with modifying security configurations?").Show()
	if err != nil {
		return false, err
	}

	return confirmed, nil
}

// HandleCopyFromOrg handles the copy-from-org functionality
func HandleCopyFromOrg(copyFromOrg string) (string, string, map[string]interface{}, string, bool, error) {
	pterm.Info.Printf("Fetching security configurations from organization '%s'...\n", copyFromOrg)

	// Check if user has access to the source organization
	status, err := api.CheckSingleOrganizationMembership(copyFromOrg)
	if err != nil {
		return "", "", nil, "", false, fmt.Errorf("failed to check membership for organization '%s': %w", copyFromOrg, err)
	}
	if !status.IsMember {
		return "", "", nil, "", false, fmt.Errorf("you are not a member of organization '%s'", copyFromOrg)
	}
	if !status.IsOwner {
		return "", "", nil, "", false, fmt.Errorf("you are a member but not an owner of organization '%s'", copyFromOrg)
	}

	// Fetch security configurations from the source organization
	configs, err := api.FetchSecurityConfigurations(copyFromOrg)
	if err != nil {
		return "", "", nil, "", false, fmt.Errorf("failed to fetch security configurations from organization '%s': %w", copyFromOrg, err)
	}

	if len(configs) == 0 {
		return "", "", nil, "", false, fmt.Errorf("no security configurations found in organization '%s'", copyFromOrg)
	}

	// Present configurations for selection
	var configOptions []string
	configMap := make(map[string]types.SecurityConfiguration)
	for _, config := range configs {
		displayName := fmt.Sprintf("%s - %s", config.Name, config.Description)
		configOptions = append(configOptions, displayName)
		configMap[displayName] = config
	}

	selectedConfig, err := pterm.DefaultInteractiveSelect.WithOptions(configOptions).Show("Select a configuration to copy")
	if err != nil {
		return "", "", nil, "", false, err
	}

	// Get the selected configuration details
	selectedConfigData := configMap[selectedConfig]

	// Get detailed configuration including settings
	configDetails, err := api.GetSecurityConfigurationDetails(copyFromOrg, selectedConfigData.ID)
	if err != nil {
		return "", "", nil, "", false, fmt.Errorf("failed to fetch configuration details: %w", err)
	}

	pterm.Success.Printf("Selected configuration '%s' from organization '%s'\n", selectedConfigData.Name, copyFromOrg)

	// Display current settings
	pterm.Info.Println("Configuration details that will be copied:")
	DisplayCurrentSettings(configDetails.Settings, configDetails.Description)
	pterm.Println()

	// Ask for attachment scope (this might be different for target organizations)
	scope, err := GetAttachmentScope()
	if err != nil {
		return "", "", nil, "", false, err
	}

	// Ask about setting as default (this might be different for target organizations)
	setAsDefault, err := GetDefaultSetting()
	if err != nil {
		return "", "", nil, "", false, err
	}

	return selectedConfigData.Name, configDetails.Description, configDetails.Settings, scope, setAsDefault, nil
}

// ConfirmApplyOperation shows operation summary and asks for confirmation for apply command
func ConfirmApplyOperation(orgs []string, configName, configDescription string, settings map[string]interface{}, scope string, setAsDefault bool) (bool, error) {
	pterm.Println()
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgYellow)).WithTextStyle(pterm.NewStyle(pterm.FgBlack)).Println("Apply Operation Summary")

	pterm.Printf("Organizations: %d\n", len(orgs))
	pterm.Printf("Configuration Name: %s\n", pterm.Yellow(configName))
	pterm.Printf("Configuration Description: %s\n", pterm.Yellow(configDescription))
	pterm.Println()

	pterm.Info.Println("Security Settings:")
	for key, value := range settings {
		valueStr := fmt.Sprintf("%v", value)
		var coloredValue string

		switch valueStr {
		case "enabled", "enforced":
			coloredValue = pterm.Green(valueStr)
		case "disabled", "unenforced":
			coloredValue = pterm.Red(valueStr)
		case "not_set":
			coloredValue = pterm.Yellow(valueStr)
		default:
			coloredValue = pterm.Yellow(valueStr)
		}

		pterm.Printf("  %s: %s\n", pterm.Cyan(key), coloredValue)
	}
	pterm.Println()

	pterm.Printf("Attachment Scope: %s\n", pterm.Magenta(scope))
	pterm.Printf("Set as Default: %s\n", pterm.Cyan(fmt.Sprintf("%t", setAsDefault)))
	pterm.Println()

	confirmed, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Proceed with applying security configuration to repositories?").Show()
	if err != nil {
		return false, err
	}

	return confirmed, nil
}
