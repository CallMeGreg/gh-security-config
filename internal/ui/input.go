package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
)

// GetEnterpriseInput prompts for enterprise slug or uses provided value
func GetEnterpriseInput(enterpriseFlag string) (string, error) {
	// If enterprise slug is provided via flag, use it
	if strings.TrimSpace(enterpriseFlag) != "" {
		return strings.TrimSpace(enterpriseFlag), nil
	}

	// Otherwise, prompt for input
	enterprise, err := pterm.DefaultInteractiveTextInput.WithDefaultText("").WithMultiLine(false).Show("Enter the enterprise slug (e.g., github)")
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(enterprise) == "" {
		return "", fmt.Errorf("enterprise slug is required")
	}

	return strings.TrimSpace(enterprise), nil
}

// GetServerURLInput prompts for GitHub Enterprise Server URL (assumes GHES since this tool is GHES-only)
func GetServerURLInput(serverURLFlag string) (string, error) {
	// If server URL is provided via flag, use it
	if strings.TrimSpace(serverURLFlag) != "" {
		return strings.TrimSpace(serverURLFlag), nil
	}

	// Since this tool is GHES-only, always prompt for server URL
	serverURL, err := pterm.DefaultInteractiveTextInput.WithDefaultText("").WithMultiLine(false).Show("Enter your GitHub Enterprise Server URL (e.g., github.company.com)")
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(serverURL) == "" {
		return "", fmt.Errorf("GitHub Enterprise Server URL is required")
	}

	return strings.TrimSpace(serverURL), nil
}

// GetDependabotAlertsAvailability prompts for Dependabot Alerts availability or uses provided value
func GetDependabotAlertsAvailability(dependabotAlertsAvailable *bool) (bool, error) {
	// If Dependabot Alerts availability is provided via flag, use it
	if dependabotAlertsAvailable != nil {
		return *dependabotAlertsAvailable, nil
	}

	// Otherwise, prompt for Dependabot Alerts availability
	pterm.Info.Println("To configure Dependabot Alerts, GitHub Connect and Dependabot must be enabled in your instance.")
	pterm.Info.Println("You can confirm this by navigating to: Enterprise settings → Settings → Code security and analysis")

	isAvailable, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Are Dependabot Alerts available in your instance?").WithDefaultValue(false).Show()
	if err != nil {
		return false, err
	}

	return isAvailable, nil
}

// GetDependabotSecurityUpdatesAvailability prompts for Dependabot Security Updates availability or uses provided value
func GetDependabotSecurityUpdatesAvailability(dependabotSecurityUpdatesAvailable *bool) (bool, error) {
	// If Dependabot Security Updates availability is provided via flag, use it
	if dependabotSecurityUpdatesAvailable != nil {
		return *dependabotSecurityUpdatesAvailable, nil
	}

	// Otherwise, prompt for Dependabot Security Updates availability
	pterm.Info.Println("To configure Dependabot Security Updates, additional setup beyond basic Dependabot may be required.")
	pterm.Info.Println("You can confirm this by navigating to: Enterprise settings → Settings → Code security and analysis")

	isAvailable, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Are Dependabot Security Updates available in your instance?").WithDefaultValue(false).Show()
	if err != nil {
		return false, err
	}

	return isAvailable, nil
}

// SetupGitHubHost sets the GH_HOST environment variable if using GitHub Enterprise Server
func SetupGitHubHost(serverURL string) {
	if serverURL != "" {
		os.Setenv("GH_HOST", serverURL)
		pterm.Info.Printf("Using GitHub Enterprise Server: %s\n", serverURL)
	}
}

// SelectOrgTargetingMethod prompts user to select an org targeting method
func SelectOrgTargetingMethod() (string, error) {
	options := []string{
		"all-orgs",
		"single-org",
		"org-list",
	}

	selection, err := pterm.DefaultInteractiveSelect.
		WithOptions(options).
		WithDefaultOption("all-orgs").
		Show("Select organization targeting method")
	if err != nil {
		return "", err
	}

	return selection, nil
}

// GetSingleOrgName prompts for a single organization name
func GetSingleOrgName() (string, error) {
	orgName, err := pterm.DefaultInteractiveTextInput.
		WithDefaultText("").
		WithMultiLine(false).
		Show("Enter organization name")
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(orgName) == "" {
		return "", fmt.Errorf("organization name is required")
	}

	return strings.TrimSpace(orgName), nil
}

// GetOrgListPath prompts for the path to a CSV file containing organizations
func GetOrgListPath() (string, error) {
	csvPath, err := pterm.DefaultInteractiveTextInput.
		WithDefaultText("organizations.csv").
		WithMultiLine(false).
		Show("Enter path to CSV file containing organization names")
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(csvPath) == "" {
		return "", fmt.Errorf("CSV file path is required")
	}

	return strings.TrimSpace(csvPath), nil
}

// GetTemplateOrgInput prompts for template organization name or uses provided value
func GetTemplateOrgInput(templateOrgFlag string) (string, error) {
	// If template org is provided via flag, use it
	if strings.TrimSpace(templateOrgFlag) != "" {
		return strings.TrimSpace(templateOrgFlag), nil
	}

	// Otherwise, prompt for input
	templateOrg, err := pterm.DefaultInteractiveTextInput.
		WithDefaultText("").
		WithMultiLine(false).
		Show("Enter the template organization name (to fetch security configurations from)")
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(templateOrg) == "" {
		return "", fmt.Errorf("template organization name is required")
	}

	return strings.TrimSpace(templateOrg), nil
}
