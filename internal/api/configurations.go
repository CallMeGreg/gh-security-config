package api

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2"
	"github.com/pterm/pterm"

	"github.com/callmegreg/gh-security-config/internal/types"
)

// FetchSecurityConfigurations retrieves all security configurations for an organization
func FetchSecurityConfigurations(org string) ([]types.SecurityConfiguration, error) {
	response, stderr, err := gh.Exec("api", "-H", "Accept: application/vnd.github+json", "-H", "X-GitHub-Api-Version: 2022-11-28", fmt.Sprintf("/orgs/%s/code-security/configurations", org))
	if err != nil {
		pterm.Error.Printf("Failed to fetch security configurations for org '%s': %v\n", org, err)
		pterm.Error.Printf("gh CLI stderr: %s\n", stderr.String())
		return nil, err
	}

	var configs []types.SecurityConfiguration
	if err := json.Unmarshal(response.Bytes(), &configs); err != nil {
		return nil, err
	}

	return configs, nil
}

// GetSecurityConfigurationDetails retrieves detailed information about a security configuration
func GetSecurityConfigurationDetails(org string, configID int) (*types.SecurityConfigurationDetails, error) {
	response, stderr, err := gh.Exec("api", "-H", "Accept: application/vnd.github+json", "-H", "X-GitHub-Api-Version: 2022-11-28", fmt.Sprintf("/orgs/%s/code-security/configurations/%d", org, configID))
	if err != nil {
		pterm.Error.Printf("Failed to fetch security configuration details for org '%s': %v\n", org, err)
		pterm.Error.Printf("gh CLI stderr: %s\n", stderr.String())
		return nil, err
	}

	var configResponse map[string]interface{}
	if err := json.Unmarshal(response.Bytes(), &configResponse); err != nil {
		return nil, err
	}

	details := &types.SecurityConfigurationDetails{
		Settings: make(map[string]interface{}),
	}

	// Extract basic info
	if id, ok := configResponse["id"].(float64); ok {
		details.ID = int(id)
	}
	if name, ok := configResponse["name"].(string); ok {
		details.Name = name
	}
	if desc, ok := configResponse["description"].(string); ok {
		details.Description = desc
	}

	// Set TargetType from API response or default to "organization"
	if targetType, ok := configResponse["target_type"].(string); ok {
		details.TargetType = targetType
	} else {
		details.TargetType = "organization"
	}

	// Extract security settings
	securitySettings := []string{
		"advanced_security", "dependabot_alerts", "dependabot_security_updates",
		"secret_scanning", "secret_scanning_push_protection",
		"secret_scanning_non_provider_patterns", "enforcement",
	}

	for _, setting := range securitySettings {
		if val, exists := configResponse[setting]; exists {
			details.Settings[setting] = val
		}
	}

	return details, nil
}

// FindConfigurationByName finds a configuration by name and returns its ID
func FindConfigurationByName(configs []types.SecurityConfiguration, name string) (int, bool) {
	for _, config := range configs {
		if config.Name == name {
			return config.ID, true
		}
	}
	return 0, false
}

// CreateSecurityConfiguration creates a new security configuration in an organization
func CreateSecurityConfiguration(org, name, description string, settings map[string]interface{}) (int, error) {
	// Build the request body
	body := map[string]interface{}{
		"name":        name,
		"description": description,
	}

	// Add all settings to the body
	for key, value := range settings {
		body[key] = value
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	// Create temporary file for the JSON body
	tmpFile, err := os.CreateTemp("", "security-config-*.json")
	if err != nil {
		return 0, err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(bodyBytes); err != nil {
		return 0, err
	}
	tmpFile.Close()

	// Execute the gh API command
	response, stderr, err := gh.Exec("api", "--method", "POST", "-H", "Accept: application/vnd.github+json", "-H", "X-GitHub-Api-Version: 2022-11-28", fmt.Sprintf("/orgs/%s/code-security/configurations", org), "--input", tmpFile.Name())
	if err != nil {
		pterm.Error.Printf("Failed to create security configuration for org '%s': %v\n", org, err)
		pterm.Error.Printf("gh CLI stderr: %s\n", stderr.String())

		// Check for 422 status code related to Dependabot unavailability
		if apiErr := parseAPIError(stderr.String(), org, settings); apiErr != nil {
			return 0, apiErr
		}

		return 0, err
	}

	var config types.SecurityConfiguration
	if err := json.Unmarshal(response.Bytes(), &config); err != nil {
		return 0, err
	}

	return config.ID, nil
}

// UpdateSecurityConfiguration updates an existing security configuration
func UpdateSecurityConfiguration(org string, configID int, name, description string, settings map[string]interface{}) error {
	// Build the request body for PATCH request
	body := map[string]interface{}{
		"name":        name,
		"description": description,
	}

	// Add all settings to the body
	for key, value := range settings {
		body[key] = value
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// Create temporary file for the JSON body
	tmpFile, err := os.CreateTemp("", "update-config-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(bodyBytes); err != nil {
		return err
	}
	tmpFile.Close()

	// Execute the gh API command with PATCH method
	_, stderr, err := gh.Exec("api", "--method", "PATCH", "-H", "Accept: application/vnd.github+json", "-H", "X-GitHub-Api-Version: 2022-11-28", fmt.Sprintf("/orgs/%s/code-security/configurations/%d", org, configID), "--input", tmpFile.Name())
	if err != nil {
		pterm.Error.Printf("Failed to update security configuration %d for org '%s': %v\n", configID, org, err)
		pterm.Error.Printf("gh CLI stderr: %s\n", stderr.String())
		return err
	}

	return nil
}

// DeleteSecurityConfiguration deletes a security configuration from an organization
func DeleteSecurityConfiguration(org string, configID int) error {
	_, stderr, err := gh.Exec("api", "--method", "DELETE", "-H", "Accept: application/vnd.github+json", "-H", "X-GitHub-Api-Version: 2022-11-28", fmt.Sprintf("/orgs/%s/code-security/configurations/%d", org, configID))
	if err != nil {
		pterm.Error.Printf("Failed to delete security configuration %d from org '%s': %v\n", configID, org, err)
		pterm.Error.Printf("gh CLI stderr: %s\n", stderr.String())
		return err
	}

	return nil
}

// AttachConfigurationToRepos attaches a security configuration to repositories
func AttachConfigurationToRepos(org string, configID int, scope string) error {
	body := map[string]interface{}{
		"scope": scope,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// Create temporary file for the JSON body
	tmpFile, err := os.CreateTemp("", "attach-config-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(bodyBytes); err != nil {
		return err
	}
	tmpFile.Close()

	_, _, err = gh.Exec("api", "--method", "POST", "-H", "Accept: application/vnd.github+json", "-H", "X-GitHub-Api-Version: 2022-11-28", fmt.Sprintf("/orgs/%s/code-security/configurations/%d/attach", org, configID), "--input", tmpFile.Name())
	return err
}

// SetConfigurationAsDefault sets a security configuration as default for new repositories
func SetConfigurationAsDefault(org string, configID int) error {
	body := map[string]interface{}{
		"default_for_new_repos": "all",
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// Create temporary file for the JSON body
	tmpFile, err := os.CreateTemp("", "default-config-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(bodyBytes); err != nil {
		return err
	}
	tmpFile.Close()

	_, _, err = gh.Exec("api", "--method", "PUT", "-H", "Accept: application/vnd.github+json", "-H", "X-GitHub-Api-Version: 2022-11-28", fmt.Sprintf("/orgs/%s/code-security/configurations/%d/defaults", org, configID), "--input", tmpFile.Name())
	return err
}

// parseAPIError checks for 422 status codes related to Dependabot unavailability
func parseAPIError(stderr string, org string, settings map[string]interface{}) error {
	if strings.Contains(stderr, "422") {
		// Check for specific Dependabot Alerts errors
		if val, hasDependabotAlerts := settings["dependabot_alerts"]; hasDependabotAlerts {
			if val != "not_set" && val != "disabled" {
				return &types.DependabotUnavailableError{
					Feature: "Dependabot Alerts",
					OrgName: org,
				}
			}
		}

		// Check for specific Dependabot Security Updates errors
		if val, hasDependabotUpdates := settings["dependabot_security_updates"]; hasDependabotUpdates {
			if val != "not_set" && val != "disabled" {
				return &types.DependabotUnavailableError{
					Feature: "Dependabot Security Updates",
					OrgName: org,
				}
			}
		}
	}
	return nil
}

// FetchEnterpriseSecurityConfigurations retrieves all security configurations for an enterprise
// This endpoint is available in GHES 3.17+
func FetchEnterpriseSecurityConfigurations(enterprise string) ([]types.SecurityConfiguration, error) {
	response, stderr, err := gh.Exec("api", "-H", "Accept: application/vnd.github+json", "-H", "X-GitHub-Api-Version: 2022-11-28", fmt.Sprintf("/enterprises/%s/code-security/configurations", enterprise))
	if err != nil {
		pterm.Error.Printf("Failed to fetch enterprise security configurations for '%s': %v\n", enterprise, err)
		pterm.Error.Printf("gh CLI stderr: %s\n", stderr.String())
		return nil, err
	}

	var configs []types.SecurityConfiguration
	if err := json.Unmarshal(response.Bytes(), &configs); err != nil {
		return nil, err
	}

	// Ensure all configs have TargetType set to "enterprise"
	for i := range configs {
		if configs[i].TargetType == "" {
			configs[i].TargetType = "enterprise"
		}
	}

	return configs, nil
}

// SupportsEnterpriseConfigurations checks if the GHES version supports enterprise-level security configurations
// Enterprise configurations are available in GHES 3.17+
func SupportsEnterpriseConfigurations(ghesVersion string) bool {
	// Parse version number
	versionFloat, err := strconv.ParseFloat(ghesVersion, 64)
	if err != nil {
		return false
	}
	return versionFloat >= 3.17
}

// GetEnterpriseSecurityConfigurationDetails retrieves detailed information about an enterprise security configuration
func GetEnterpriseSecurityConfigurationDetails(enterprise string, configID int) (*types.SecurityConfigurationDetails, error) {
	response, stderr, err := gh.Exec("api", "-H", "Accept: application/vnd.github+json", "-H", "X-GitHub-Api-Version: 2022-11-28", fmt.Sprintf("/enterprises/%s/code-security/configurations/%d", enterprise, configID))
	if err != nil {
		pterm.Error.Printf("Failed to fetch enterprise security configuration details: %v\n", err)
		pterm.Error.Printf("gh CLI stderr: %s\n", stderr.String())
		return nil, err
	}

	var configResponse map[string]interface{}
	if err := json.Unmarshal(response.Bytes(), &configResponse); err != nil {
		return nil, err
	}

	details := &types.SecurityConfigurationDetails{
		Settings:   make(map[string]interface{}),
		TargetType: "enterprise",
	}

	// Extract basic info
	if id, ok := configResponse["id"].(float64); ok {
		details.ID = int(id)
	}
	if name, ok := configResponse["name"].(string); ok {
		details.Name = name
	}
	if desc, ok := configResponse["description"].(string); ok {
		details.Description = desc
	}

	// Extract security settings
	securitySettings := []string{
		"advanced_security", "dependabot_alerts", "dependabot_security_updates",
		"secret_scanning", "secret_scanning_push_protection",
		"secret_scanning_non_provider_patterns", "enforcement",
	}

	for _, setting := range securitySettings {
		if val, exists := configResponse[setting]; exists {
			details.Settings[setting] = val
		}
	}

	return details, nil
}
