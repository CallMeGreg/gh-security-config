package types

import "fmt"

// ConfigurationExistsError represents an error when a security configuration already exists
type ConfigurationExistsError struct {
	ConfigName string
	OrgName    string
}

func (e *ConfigurationExistsError) Error() string {
	return fmt.Sprintf("configuration '%s' already exists in organization '%s'", e.ConfigName, e.OrgName)
}

// DependabotUnavailableError represents an error when Dependabot features are not available
type DependabotUnavailableError struct {
	Feature string
	OrgName string
}

func (e *DependabotUnavailableError) Error() string {
	return fmt.Sprintf("Dependabot %s is not available for organization '%s'. This feature may not be enabled on your GitHub Enterprise Server instance", e.Feature, e.OrgName)
}
