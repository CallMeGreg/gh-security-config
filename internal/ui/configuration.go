package ui

import (
	"fmt"
	"strings"

	"github.com/pterm/pterm"
)

// GetSecurityConfigInput prompts for security configuration name and description.
// If nameOverride or descriptionOverride are non-empty, they are used instead of prompting.
func GetSecurityConfigInput(nameOverride, descriptionOverride string) (string, string, error) {
	var name string
	if strings.TrimSpace(nameOverride) != "" {
		name = strings.TrimSpace(nameOverride)
	} else {
		n, err := pterm.DefaultInteractiveTextInput.WithDefaultText("Security Configuration").WithMultiLine(false).Show("Enter security configuration name")
		if err != nil {
			return "", "", err
		}
		name = strings.TrimSpace(n)
	}
	if name == "" {
		return "", "", fmt.Errorf("configuration name is required")
	}

	var description string
	if strings.TrimSpace(descriptionOverride) != "" {
		description = strings.TrimSpace(descriptionOverride)
	} else {
		d, err := pterm.DefaultInteractiveTextInput.WithDefaultText("Security configuration applied across enterprise organizations").WithMultiLine(false).Show("Enter security configuration description")
		if err != nil {
			return "", "", err
		}
		description = strings.TrimSpace(d)
	}
	if description == "" {
		return "", "", fmt.Errorf("configuration description is required")
	}

	return name, description, nil
}

// SecuritySettingOverrides holds optional pre-supplied values for security settings.
// Any field left empty will fall back to interactive prompting.
type SecuritySettingOverrides struct {
	AdvancedSecurity                  string
	DependabotAlerts                  string
	DependabotSecurityUpdates         string
	SecretScanning                    string
	SecretScanningPushProtection      string
	SecretScanningNonProviderPatterns string
	Enforcement                       string
}

// selectWithOverride validates an override (if provided) against allowed options.
// If the override is empty, it prompts the user with the given label and default.
func selectWithOverride(label, override string, options []string, defaultOption string) (string, error) {
	if override != "" {
		for _, o := range options {
			if o == override {
				return override, nil
			}
		}
		return "", fmt.Errorf("invalid value %q for %s (must be one of: %s)", override, label, strings.Join(options, ", "))
	}
	return pterm.DefaultInteractiveSelect.WithOptions(options).WithDefaultOption(defaultOption).Show(label)
}

// GetSecuritySettings prompts for security settings configuration. Any non-empty field on
// overrides is used directly without prompting the user.
func GetSecuritySettings(overrides SecuritySettingOverrides, dependabotAlertsAvailable bool, dependabotSecurityUpdatesAvailable bool) (map[string]interface{}, error) {
	settings := make(map[string]interface{})

	// Only show the header if at least one setting will actually be prompted for
	needsPrompt := overrides.AdvancedSecurity == "" ||
		(dependabotAlertsAvailable && overrides.DependabotAlerts == "") ||
		(dependabotSecurityUpdatesAvailable && overrides.DependabotSecurityUpdates == "") ||
		overrides.SecretScanning == "" ||
		overrides.SecretScanningPushProtection == "" ||
		overrides.SecretScanningNonProviderPatterns == "" ||
		overrides.Enforcement == ""
	if needsPrompt {
		pterm.Info.Println("Configure security settings:")
	}

	// Advanced Security
	advancedSecurity, err := selectWithOverride("GitHub Advanced Security", overrides.AdvancedSecurity, []string{"enabled", "disabled"}, "enabled")
	if err != nil {
		return nil, err
	}
	settings["advanced_security"] = advancedSecurity

	// Dependabot Alerts (only if available)
	if dependabotAlertsAvailable {
		dependabotAlerts, err := selectWithOverride("Dependabot Alerts", overrides.DependabotAlerts, []string{"enabled", "disabled", "not_set"}, "not_set")
		if err != nil {
			return nil, err
		}
		settings["dependabot_alerts"] = dependabotAlerts
	}

	// Dependabot Security Updates (only if available)
	if dependabotSecurityUpdatesAvailable {
		dependabotSecurityUpdates, err := selectWithOverride("Dependabot Security Updates", overrides.DependabotSecurityUpdates, []string{"enabled", "disabled", "not_set"}, "not_set")
		if err != nil {
			return nil, err
		}
		settings["dependabot_security_updates"] = dependabotSecurityUpdates
	}

	// Secret Scanning
	secretScanning, err := selectWithOverride("Secret Scanning", overrides.SecretScanning, []string{"enabled", "disabled", "not_set"}, "enabled")
	if err != nil {
		return nil, err
	}
	settings["secret_scanning"] = secretScanning

	// Secret Scanning Push Protection
	pushProtection, err := selectWithOverride("Secret Scanning Push Protection", overrides.SecretScanningPushProtection, []string{"enabled", "disabled", "not_set"}, "enabled")
	if err != nil {
		return nil, err
	}
	settings["secret_scanning_push_protection"] = pushProtection

	// Secret Scanning Non-Provider Patterns
	nonProviderPatterns, err := selectWithOverride("Secret Scanning Non-Provider Patterns", overrides.SecretScanningNonProviderPatterns, []string{"enabled", "disabled", "not_set"}, "not_set")
	if err != nil {
		return nil, err
	}
	settings["secret_scanning_non_provider_patterns"] = nonProviderPatterns

	// Enforcement
	enforcement, err := selectWithOverride("Enforcement Status", overrides.Enforcement, []string{"enforced", "unenforced"}, "enforced")
	if err != nil {
		return nil, err
	}
	settings["enforcement"] = enforcement

	return settings, nil
}

// GetAttachmentScope prompts for repository attachment scope. If override is non-empty,
// it is validated and used directly.
func GetAttachmentScope(override string) (string, error) {
	options := []string{"all", "public", "private_or_internal", "none"}
	if override != "" {
		for _, o := range options {
			if o == override {
				return override, nil
			}
		}
		return "", fmt.Errorf("invalid value %q for scope (must be one of: %s)", override, strings.Join(options, ", "))
	}
	scope, err := pterm.DefaultInteractiveSelect.WithOptions(options).WithDefaultOption("all").Show("Select repositories to attach configuration to")
	if err != nil {
		return "", err
	}

	return scope, nil
}

// GetDefaultSetting prompts whether to set configuration as default. If override is non-nil,
// its value is used directly.
func GetDefaultSetting(override *bool) (bool, error) {
	if override != nil {
		return *override, nil
	}
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

// GetUpdatedName prompts for updated configuration name. If override is non-empty, it is used.
// To explicitly mean "keep current", pass currentName as the override (the caller handles that).
func GetUpdatedName(currentName, override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		newName := strings.TrimSpace(override)
		return newName, nil
	}
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

// GetUpdatedDescription prompts for updated description. If override is non-empty, it is used
// directly without prompting; otherwise the user is prompted and the current value is offered
// as the default (pressing Enter keeps the current description).
func GetUpdatedDescription(currentDescription, override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		return strings.TrimSpace(override), nil
	}
	newDescription, err := pterm.DefaultInteractiveTextInput.WithDefaultText(currentDescription).WithMultiLine(false).Show("Enter updated security configuration description")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(newDescription), nil
}

// GetSecuritySettingsForUpdate prompts for updated security settings. Any non-empty override
// on overrides is used directly instead of prompting. Unspecified settings default to keeping
// the current value.
func GetSecuritySettingsForUpdate(currentSettings map[string]interface{}, overrides SecuritySettingOverrides, dependabotAlertsAvailable bool, dependabotSecurityUpdatesAvailable bool) (map[string]interface{}, error) {
	newSettings := make(map[string]interface{})

	settingsConfig := []struct {
		key                          string
		description                  string
		options                      []string
		defaultValue                 string
		override                     string
		requiresDependabotAlerts     bool
		requiresDependabotSecUpdates bool
	}{
		{"advanced_security", "GitHub Advanced Security", []string{"enabled", "disabled"}, "enabled", overrides.AdvancedSecurity, false, false},
		{"dependabot_alerts", "Dependabot Alerts", []string{"enabled", "disabled", "not_set"}, "not_set", overrides.DependabotAlerts, true, false},
		{"dependabot_security_updates", "Dependabot Security Updates", []string{"enabled", "disabled", "not_set"}, "not_set", overrides.DependabotSecurityUpdates, false, true},
		{"secret_scanning", "Secret Scanning", []string{"enabled", "disabled", "not_set"}, "enabled", overrides.SecretScanning, false, false},
		{"secret_scanning_push_protection", "Secret Scanning Push Protection", []string{"enabled", "disabled", "not_set"}, "enabled", overrides.SecretScanningPushProtection, false, false},
		{"secret_scanning_non_provider_patterns", "Secret Scanning Non-Provider Patterns", []string{"enabled", "disabled", "not_set"}, "not_set", overrides.SecretScanningNonProviderPatterns, false, false},
		{"enforcement", "Enforcement Status", []string{"enforced", "unenforced"}, "enforced", overrides.Enforcement, false, false},
	}

	// Determine if we will prompt for anything (to decide whether to show the header)
	willPrompt := false
	for _, c := range settingsConfig {
		if c.requiresDependabotAlerts && !dependabotAlertsAvailable {
			continue
		}
		if c.requiresDependabotSecUpdates && !dependabotSecurityUpdatesAvailable {
			continue
		}
		if c.override == "" {
			willPrompt = true
			break
		}
	}
	if willPrompt {
		pterm.Info.Println("Update security settings (press Enter to keep current value):")
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

		// If an override is provided via flag, validate and use it
		if config.override != "" {
			valid := false
			for _, o := range config.options {
				if o == config.override {
					valid = true
					break
				}
			}
			if !valid {
				return nil, fmt.Errorf("invalid value %q for %s (must be one of: %s)", config.override, config.description, strings.Join(config.options, ", "))
			}
			newSettings[config.key] = config.override
			continue
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

// SelectConfigurationFromList prompts user to select a configuration from a list.
// If override is non-empty, the matching config is returned directly. If configSource is
// also provided (either "organization" or "enterprise"), it disambiguates when the same name
// exists in both lists.
// Returns the configuration name and target type (organization or enterprise)
func SelectConfigurationFromList(orgConfigs, enterpriseConfigs []string, override, configSource string) (string, string, error) {
	if override != "" {
		return resolveConfigOverride(orgConfigs, enterpriseConfigs, override, configSource)
	}
	return selectConfiguration(orgConfigs, enterpriseConfigs, "Select a security configuration to apply")
}

// SelectConfigurationForDeletion prompts user to select a configuration to delete.
// If override is non-empty and matches one of the configs, it is returned directly.
// Returns the configuration name
func SelectConfigurationForDeletion(orgConfigs []string, override string) (string, error) {
	if override != "" {
		return resolveNameOverride(orgConfigs, override, "delete")
	}
	return selectFromList(orgConfigs, "Select a security configuration to delete")
}

// SelectConfigurationForModification prompts user to select a configuration to modify.
// If override is non-empty and matches one of the configs, it is returned directly.
// Returns the configuration name
func SelectConfigurationForModification(orgConfigs []string, override string) (string, error) {
	if override != "" {
		return resolveNameOverride(orgConfigs, override, "modify")
	}
	return selectFromList(orgConfigs, "Select a security configuration to modify")
}

// resolveNameOverride validates that the supplied override matches one of the available configs.
func resolveNameOverride(configs []string, override, verb string) (string, error) {
	for _, c := range configs {
		if c == override {
			return override, nil
		}
	}
	return "", fmt.Errorf("configuration %q not found in the list of configurations available to %s", override, verb)
}

// resolveConfigOverride disambiguates between org and enterprise configs given an override name
// and optional configSource ("organization" or "enterprise").
func resolveConfigOverride(orgConfigs, enterpriseConfigs []string, override, configSource string) (string, string, error) {
	inOrg := contains(orgConfigs, override)
	inEnterprise := contains(enterpriseConfigs, override)

	switch configSource {
	case "organization":
		if !inOrg {
			return "", "", fmt.Errorf("organization configuration %q not found in template organization", override)
		}
		return override, "organization", nil
	case "enterprise":
		if !inEnterprise {
			return "", "", fmt.Errorf("enterprise configuration %q not found", override)
		}
		return override, "enterprise", nil
	case "":
		if inOrg && inEnterprise {
			return "", "", fmt.Errorf("configuration name %q exists at both organization and enterprise level; specify --config-source (organization|enterprise) to disambiguate", override)
		}
		if inOrg {
			return override, "organization", nil
		}
		if inEnterprise {
			return override, "enterprise", nil
		}
		return "", "", fmt.Errorf("configuration %q not found at organization or enterprise level", override)
	default:
		return "", "", fmt.Errorf("invalid value %q for --config-source (must be 'organization' or 'enterprise')", configSource)
	}
}

func contains(list []string, target string) bool {
	for _, s := range list {
		if s == target {
			return true
		}
	}
	return false
}

// selectFromList is a shared helper for single-list configuration selection prompts
func selectFromList(configs []string, prompt string) (string, error) {
	if len(configs) == 0 {
		return "", fmt.Errorf("no configurations available")
	}

	selection, err := pterm.DefaultInteractiveSelect.WithOptions(configs).Show(prompt)
	if err != nil {
		return "", err
	}

	return selection, nil
}

// selectConfiguration is a shared helper for configuration selection prompts
func selectConfiguration(orgConfigs, enterpriseConfigs []string, prompt string) (string, string, error) {
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

	selection, err := pterm.DefaultInteractiveSelect.WithOptions(options).Show(prompt)
	if err != nil {
		return "", "", err
	}

	config := configMap[selection]
	return config.name, config.targetType, nil
}

// GetAttachmentScopeForApplication prompts for repository attachment scope (without 'none' option).
// If override is non-empty, it is validated and used directly.
func GetAttachmentScopeForApplication(override string) (string, error) {
	options := []string{"all", "public", "private_or_internal"}
	if override != "" {
		for _, o := range options {
			if o == override {
				return override, nil
			}
		}
		return "", fmt.Errorf("invalid value %q for scope (must be one of: %s)", override, strings.Join(options, ", "))
	}
	scope, err := pterm.DefaultInteractiveSelect.WithOptions(options).WithDefaultOption("all").Show("Select repositories to attach configuration to")
	if err != nil {
		return "", err
	}

	return scope, nil
}
