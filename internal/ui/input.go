package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
)

// GetEnterpriseInput prompts for enterprise slug
func GetEnterpriseInput() (string, error) {
	enterprise, err := pterm.DefaultInteractiveTextInput.WithDefaultText("").WithMultiLine(false).Show("Enter the enterprise slug (e.g., github)")
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(enterprise) == "" {
		return "", fmt.Errorf("enterprise slug is required")
	}

	return strings.TrimSpace(enterprise), nil
}

// GetServerURLInput prompts for GitHub Enterprise Server URL if needed
func GetServerURLInput() (string, error) {
	isGHES, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Are you using GitHub Enterprise Server (not GitHub.com)?").WithDefaultValue(true).Show()
	if err != nil {
		return "", err
	}

	if !isGHES {
		return "", nil
	}

	serverURL, err := pterm.DefaultInteractiveTextInput.WithDefaultText("").WithMultiLine(false).Show("Enter your GitHub Enterprise Server URL (e.g., github.company.com)")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(serverURL), nil
}

// SetupGitHubHost sets the GH_HOST environment variable if using GitHub Enterprise Server
func SetupGitHubHost(serverURL string) {
	if serverURL != "" {
		os.Setenv("GH_HOST", serverURL)
		pterm.Info.Printf("Using GitHub Enterprise Server: %s\n", serverURL)
	}
}
