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
