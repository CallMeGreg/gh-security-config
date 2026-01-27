package ui

import (
	"fmt"
	"strings"

	"github.com/pterm/pterm"
)

// GetSecurityConfigInput prompts for security configuration name and description
func GetSecurityConfigInput() (string, string, error) {
	name, err := pterm.DefaultInteractiveTextInput.WithDefaultText("Enterprise Security Configuration").WithMultiLine(false).Show("Enter security configuration name")
	if err != nil {
		return "", "", err
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return "", "", fmt.Errorf("configuration name is required")
	}

	description, err := pterm.DefaultInteractiveTextInput.WithDefaultText("Security configuration applied across enterprise organizations").WithMultiLine(false).Show("Enter security configuration description")
	if err != nil {
		return "", "", err
	}

	description = strings.TrimSpace(description)
	if description == "" {
		return "", "", fmt.Errorf("configuration description is required")
	}

	return name, description, nil
}

// GetSecuritySettings prompts for security settings configuration
func GetSecuritySettings(dependabotAlertsAvailable bool, dependabotSecurityUpdatesAvailable bool) (map[string]interface{}, error) {
	settings := make(map[string]interface{})

	pterm.Info.Println("Configure security settings:")

	// Advanced Security
	advancedSecurity, err := pterm.DefaultInteractiveSelect.WithOptions([]string{"enabled", "disabled"}).WithDefaultOption("enabled").Show("GitHub Advanced Security")
	if err != nil {
		return nil, err
	}
	settings["advanced_security"] = advancedSecurity

	// Dependabot Alerts (only if available)
	if dependabotAlertsAvailable {
		dependabotAlerts, err := pterm.DefaultInteractiveSelect.WithOptions([]string{"enabled", "disabled", "not_set"}).WithDefaultOption("not_set").Show("Dependabot Alerts")
		if err != nil {
			return nil, err
		}
		settings["dependabot_alerts"] = dependabotAlerts
	}

	// Dependabot Security Updates (only if available)
	if dependabotSecurityUpdatesAvailable {
		dependabotSecurityUpdates, err := pterm.DefaultInteractiveSelect.WithOptions([]string{"enabled", "disabled", "not_set"}).WithDefaultOption("not_set").Show("Dependabot Security Updates")
		if err != nil {
			return nil, err
		}
		settings["dependabot_security_updates"] = dependabotSecurityUpdates
	}

	// Secret Scanning
	secretScanning, err := pterm.DefaultInteractiveSelect.WithOptions([]string{"enabled", "disabled", "not_set"}).WithDefaultOption("enabled").Show("Secret Scanning")
	if err != nil {
		return nil, err
	}
	settings["secret_scanning"] = secretScanning

	// Secret Scanning Push Protection
	pushProtection, err := pterm.DefaultInteractiveSelect.WithOptions([]string{"enabled", "disabled", "not_set"}).WithDefaultOption("enabled").Show("Secret Scanning Push Protection")
	if err != nil {
		return nil, err
	}
	settings["secret_scanning_push_protection"] = pushProtection

	// Secret Scanning Non-Provider Patterns
	nonProviderPatterns, err := pterm.DefaultInteractiveSelect.WithOptions([]string{"enabled", "disabled", "not_set"}).WithDefaultOption("not_set").Show("Secret Scanning Non-Provider Patterns")
	if err != nil {
		return nil, err
	}
	settings["secret_scanning_non_provider_patterns"] = nonProviderPatterns

	// Enforcement
	enforcement, err := pterm.DefaultInteractiveSelect.WithOptions([]string{"enforced", "unenforced"}).WithDefaultOption("enforced").Show("Enforcement Status")
	if err != nil {
		return nil, err
	}
	settings["enforcement"] = enforcement

	return settings, nil
}

// GetAttachmentScope prompts for repository attachment scope
func GetAttachmentScope() (string, error) {
	scope, err := pterm.DefaultInteractiveSelect.WithOptions([]string{
		"all",
		"public",
		"private_or_internal",
		"none",
	}).WithDefaultOption("all").Show("Select repositories to attach configuration to")
	if err != nil {
		return "", err
	}

	return scope, nil
}

// GetDefaultSetting prompts whether to set configuration as default
func GetDefaultSetting() (bool, error) {
	setDefault, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Set this configuration as default for new repositories?").Show()
	if err != nil {
		return false, err
	}

	return setDefault, nil
}

// GetConfigNameForDeletion prompts for configuration name to delete
func GetConfigNameForDeletion() (string, error) {
	configName, err := pterm.DefaultInteractiveTextInput.WithDefaultText("").WithMultiLine(false).Show("Enter the name of the security configuration to delete")
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(configName) == "" {
		return "", fmt.Errorf("configuration name is required")
	}

	return strings.TrimSpace(configName), nil
}

// GetConfigNameForModification prompts for configuration name to modify
func GetConfigNameForModification() (string, error) {
	configName, err := pterm.DefaultInteractiveTextInput.WithDefaultText("").WithMultiLine(false).Show("Enter the name of the security configuration to modify")
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(configName) == "" {
		return "", fmt.Errorf("configuration name is required")
	}

	return strings.TrimSpace(configName), nil
}

// GetUpdatedName prompts for updated configuration name
func GetUpdatedName(currentName string) (string, error) {
	newName, err := pterm.DefaultInteractiveTextInput.WithDefaultText(currentName).WithMultiLine(false).Show("Enter updated security configuration name")
	if err != nil {
		return "", err
	}

	newName = strings.TrimSpace(newName)
	if newName == "" {
		return "", fmt.Errorf("configuration name is required")
	}

	return newName, nil
}

// GetUpdatedDescription prompts for updated description
func GetUpdatedDescription(currentDescription string) (string, error) {
	newDescription, err := pterm.DefaultInteractiveTextInput.WithDefaultText(currentDescription).WithMultiLine(false).Show("Enter updated security configuration description")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(newDescription), nil
}

// GetSecuritySettingsForUpdate prompts for updated security settings
func GetSecuritySettingsForUpdate(currentSettings map[string]interface{}, dependabotAlertsAvailable bool, dependabotSecurityUpdatesAvailable bool) (map[string]interface{}, error) {
	newSettings := make(map[string]interface{})

	pterm.Info.Println("Update security settings (press Enter to keep current value):")

	settingsConfig := []struct {
		key                          string
		description                  string
		options                      []string
		defaultValue                 string
		requiresDependabotAlerts     bool
		requiresDependabotSecUpdates bool
	}{
		{"advanced_security", "GitHub Advanced Security", []string{"enabled", "disabled"}, "enabled", false, false},
		{"dependabot_alerts", "Dependabot Alerts", []string{"enabled", "disabled", "not_set"}, "not_set", true, false},
		{"dependabot_security_updates", "Dependabot Security Updates", []string{"enabled", "disabled", "not_set"}, "not_set", false, true},
		{"secret_scanning", "Secret Scanning", []string{"enabled", "disabled", "not_set"}, "enabled", false, false},
		{"secret_scanning_push_protection", "Secret Scanning Push Protection", []string{"enabled", "disabled", "not_set"}, "enabled", false, false},
		{"secret_scanning_non_provider_patterns", "Secret Scanning Non-Provider Patterns", []string{"enabled", "disabled", "not_set"}, "not_set", false, false},
		{"enforcement", "Enforcement Status", []string{"enforced", "unenforced"}, "enforced", false, false},
	}

	for _, config := range settingsConfig {
		// Skip Dependabot settings if not available
		if config.requiresDependabotAlerts && !dependabotAlertsAvailable {
			continue
		}
		if config.requiresDependabotSecUpdates && !dependabotSecurityUpdatesAvailable {
			continue
		}

		currentValue := "not_set"
		if val, exists := currentSettings[config.key]; exists {
			currentValue = fmt.Sprintf("%v", val)
		}

		// Add option to keep current value
		options := append([]string{fmt.Sprintf("Keep current (%s)", currentValue)}, config.options...)

		selection, err := pterm.DefaultInteractiveSelect.WithOptions(options).WithDefaultOption(options[0]).Show(config.description)
		if err != nil {
			return nil, err
		}

		// If user chose to keep current value, use the current value
		if strings.HasPrefix(selection, "Keep current") {
			newSettings[config.key] = currentValue
		} else {
			newSettings[config.key] = selection
		}
	}

	return newSettings, nil
}

// GetConfigNameForApplication prompts for configuration name to apply
func GetConfigNameForApplication() (string, error) {
	configName, err := pterm.DefaultInteractiveTextInput.WithDefaultText("").WithMultiLine(false).Show("Enter the name of the security configuration to apply")
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(configName) == "" {
		return "", fmt.Errorf("configuration name is required")
	}

	return strings.TrimSpace(configName), nil
}

// SelectConfigurationFromList prompts user to select a configuration from a list
// Returns the configuration name and target type (organization or enterprise)
func SelectConfigurationFromList(orgConfigs, enterpriseConfigs []string) (string, string, error) {
	if len(orgConfigs) == 0 && len(enterpriseConfigs) == 0 {
		return "", "", fmt.Errorf("no configurations available")
	}

	// Build options list with prefixes to indicate source
	// Organization configs first as they are more commonly used
	var options []string
	configMap := make(map[string]struct {
		name       string
		targetType string
	})

	for _, name := range orgConfigs {
		option := fmt.Sprintf("[Organization] %s", name)
		options = append(options, option)
		configMap[option] = struct {
			name       string
			targetType string
		}{name, "organization"}
	}

	for _, name := range enterpriseConfigs {
		option := fmt.Sprintf("[Enterprise] %s", name)
		options = append(options, option)
		configMap[option] = struct {
			name       string
			targetType string
		}{name, "enterprise"}
	}

	selection, err := pterm.DefaultInteractiveSelect.WithOptions(options).Show("Select a security configuration to apply")
	if err != nil {
		return "", "", err
	}

	config := configMap[selection]
	return config.name, config.targetType, nil
}

// GetAttachmentScopeForApplication prompts for repository attachment scope (without 'none' option)
func GetAttachmentScopeForApplication() (string, error) {
	scope, err := pterm.DefaultInteractiveSelect.WithOptions([]string{
		"all",
		"public",
		"private_or_internal",
	}).WithDefaultOption("all").Show("Select repositories to attach configuration to")
	if err != nil {
		return "", err
	}

	return scope, nil
}
