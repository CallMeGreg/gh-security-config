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
		"template-org",
		"org",
		"org-list",
		"all-orgs",
		"copy-from-org",
		"config-name",
		"config-description",
		"new-name",
		"new-description",
		"config-source",
		"advanced-security",
		"dependabot-alerts",
		"dependabot-security-updates",
		"secret-scanning",
		"secret-scanning-push-protection",
		"secret-scanning-non-provider-patterns",
		"enforcement",
		"scope",
		"set-as-default",
		"dependabot-alerts-available",
		"dependabot-security-updates-available",
		"concurrency",
		"delay",
		"log-level",
		"skip-confirmation-message",
		"overwrite",
	}

	for _, flagName := range flagOrder {
		if value, exists := flags[flagName]; exists && value != nil {
			switch v := value.(type) {
			case string:
				if v != "" {
					// Only include log-level if it's not the default
					if flagName == "log-level" && v == "warning" {
						continue
					}
					parts = append(parts, fmt.Sprintf("--%s %s", flagName, quoteIfNeeded(v)))
				}
			case bool:
				if v {
					// Boolean flags don't need a value
					parts = append(parts, fmt.Sprintf("--%s", flagName))
				}
			case int:
				if (flagName == "concurrency" && v != 1) || (flagName == "delay" && v != 0) {
					// Only include concurrency if it's not the default (1) or delay if it's not default (0)
					parts = append(parts, fmt.Sprintf("--%s %d", flagName, v))
				}
			}
		}
	}

	return strings.Join(parts, " ")
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
	pterm.Println(pterm.NewStyle(pterm.FgWhite).Sprint("> ") + pterm.NewStyle(pterm.FgLightGreen).Sprint(command))
	pterm.Println()
}
