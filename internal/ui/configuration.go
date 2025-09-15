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

	description, err := pterm.DefaultInteractiveTextInput.WithDefaultText("Security configuration applied across enterprise organizations").WithMultiLine(false).Show("Enter security configuration description")
	if err != nil {
		return "", "", err
	}

	return strings.TrimSpace(name), strings.TrimSpace(description), nil
}

// GetSecuritySettings prompts for security settings configuration
func GetSecuritySettings() (map[string]interface{}, error) {
	settings := make(map[string]interface{})

	pterm.Info.Println("Configure security settings:")

	// Advanced Security
	advancedSecurity, err := pterm.DefaultInteractiveSelect.WithOptions([]string{"enabled", "disabled"}).WithDefaultOption("enabled").Show("GitHub Advanced Security")
	if err != nil {
		return nil, err
	}
	settings["advanced_security"] = advancedSecurity

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
	nonProviderPatterns, err := pterm.DefaultInteractiveSelect.WithOptions([]string{"enabled", "disabled", "not_set"}).WithDefaultOption("disabled").Show("Secret Scanning Non-Provider Patterns")
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

// GetUpdatedDescription prompts for updated description
func GetUpdatedDescription(currentDescription string) (string, error) {
	newDescription, err := pterm.DefaultInteractiveTextInput.WithDefaultText(currentDescription).WithMultiLine(false).Show("Enter updated security configuration description")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(newDescription), nil
}

// GetSecuritySettingsForUpdate prompts for updated security settings
func GetSecuritySettingsForUpdate(currentSettings map[string]interface{}) (map[string]interface{}, error) {
	newSettings := make(map[string]interface{})

	pterm.Info.Println("Update security settings (press Enter to keep current value):")

	settingsConfig := []struct {
		key          string
		description  string
		options      []string
		defaultValue string
	}{
		{"advanced_security", "GitHub Advanced Security", []string{"enabled", "disabled"}, "enabled"},
		{"secret_scanning", "Secret Scanning", []string{"enabled", "disabled", "not_set"}, "enabled"},
		{"secret_scanning_push_protection", "Secret Scanning Push Protection", []string{"enabled", "disabled", "not_set"}, "enabled"},
		{"secret_scanning_non_provider_patterns", "Secret Scanning Non-Provider Patterns", []string{"enabled", "disabled", "not_set"}, "disabled"},
		{"enforcement", "Enforcement Status", []string{"enforced", "unenforced"}, "enforced"},
	}

	for _, config := range settingsConfig {
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
