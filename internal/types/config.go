package types

// SecurityConfiguration represents a GitHub security configuration
type SecurityConfiguration struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// SecurityConfigurationDetails represents detailed security configuration information
type SecurityConfigurationDetails struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Settings    map[string]interface{} `json:"-"` // Will be populated separately
}

// ProcessingResult represents the result of processing a single organization
type ProcessingResult struct {
	Organization string
	Success      bool
	Skipped      bool
	Error        error
}
