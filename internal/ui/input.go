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

// GetDependabotAvailability prompts for Dependabot availability or uses provided value
func GetDependabotAvailability(dependabotAvailable *bool) (bool, error) {
	// If Dependabot availability is provided via flag, use it
	if dependabotAvailable != nil {
		return *dependabotAvailable, nil
	}

	// Otherwise, prompt for Dependabot availability
	pterm.Info.Println("To configure Dependabot settings, GitHub Connect and Dependabot must be enabled in your instance.")
	pterm.Info.Println("You can confirm this by navigating to: Enterprise settings → GitHub Connect → Dependabot")

	isAvailable, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Is Dependabot available in your instance?").WithDefaultValue(false).Show()
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
