package utils

import (
	"fmt"
	"strings"

	"github.com/pterm/pterm"
)

// BuildReplicationCommand creates a command string that can be used to replicate the same action
func BuildReplicationCommand(command string, flags map[string]interface{}) string {
	var parts []string
	parts = append(parts, "gh security-config", command)

	// Add flags in a consistent order
	flagOrder := []string{
		"enterprise-slug",
		"github-enterprise-server-url",
		"org",
		"org-list",
		"all-orgs",
		"copy-from-org",
		"force",
		"dependabot-alerts-available",
		"dependabot-security-updates-available",
		"concurrency",
		"delay",
	}

	for _, flagName := range flagOrder {
		if value, exists := flags[flagName]; exists && value != nil {
			switch v := value.(type) {
			case string:
				if v != "" {
					// Determine the short flag if available
					shortFlag := getShortFlag(flagName)
					if shortFlag != "" {
						parts = append(parts, fmt.Sprintf("-%s %s", shortFlag, quoteIfNeeded(v)))
					} else {
						parts = append(parts, fmt.Sprintf("--%s %s", flagName, quoteIfNeeded(v)))
					}
				}
			case bool:
				if v {
					// Boolean flags don't need a value
					shortFlag := getShortFlag(flagName)
					if shortFlag != "" {
						parts = append(parts, fmt.Sprintf("-%s", shortFlag))
					} else {
						parts = append(parts, fmt.Sprintf("--%s", flagName))
					}
				}
			case int:
				if v > 0 && (flagName == "concurrency" && v != 1 || flagName == "delay" && v != 0) {
					// Only include concurrency if it's not the default (1) or delay if it's not default (0)
					shortFlag := getShortFlag(flagName)
					if shortFlag != "" {
						parts = append(parts, fmt.Sprintf("-%s %d", shortFlag, v))
					} else {
						parts = append(parts, fmt.Sprintf("--%s %d", flagName, v))
					}
				}
			}
		}
	}

	return strings.Join(parts, " ")
}

// getShortFlag returns the short version of a flag if it exists
func getShortFlag(flagName string) string {
	shortFlags := map[string]string{
		"org-list":                                "l",
		"concurrency":                             "c",
		"delay":                                   "d",
		"enterprise-slug":                         "e",
		"github-enterprise-server-url":            "u",
		"dependabot-alerts-available":             "a",
		"dependabot-security-updates-available":   "s",
		"copy-from-org":                           "o",
		"force":                                   "f",
	}
	return shortFlags[flagName]
}

// quoteIfNeeded adds quotes around a string if it contains spaces
func quoteIfNeeded(s string) string {
	if strings.Contains(s, " ") {
		return fmt.Sprintf("\"%s\"", s)
	}
	return s
}

// ShowReplicationCommand displays the replication command to the user
func ShowReplicationCommand(command string) {
	pterm.Println()
	pterm.Info.Println("To replicate this operation, use the following command:")
	pterm.Println()
	
	// Use a box to highlight the command
	boxedCommand := pterm.DefaultBox.
		WithTitle("Replication Command").
		WithTitleTopCenter().
		WithRightPadding(2).
		WithLeftPadding(2).
		WithBoxStyle(pterm.NewStyle(pterm.FgCyan)).
		Sprint(command)
	
	pterm.Println(boxedCommand)
}
