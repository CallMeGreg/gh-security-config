package utils

import (
	"strings"
	"testing"
)

func TestBuildReplicationCommand(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		flags    map[string]interface{}
		expected []string // Expected substrings in the command
	}{
		{
			name:    "Generate with all-orgs and basic flags",
			command: "generate",
			flags: map[string]interface{}{
				"enterprise-slug":              "my-enterprise",
				"github-enterprise-server-url": "github.company.com",
				"all-orgs":                     true,
				"dependabot-alerts-available":  "true",
				"dependabot-security-updates-available": "false",
			},
			expected: []string{
				"gh security-config generate",
				"-e my-enterprise",
				"-u github.company.com",
				"--all-orgs",
				"-a true",
				"-s false",
			},
		},
		{
			name:    "Generate with org flag",
			command: "generate",
			flags: map[string]interface{}{
				"enterprise-slug":              "my-enterprise",
				"github-enterprise-server-url": "github.company.com",
				"org":                          "test-org",
			},
			expected: []string{
				"gh security-config generate",
				"-e my-enterprise",
				"-u github.company.com",
				"--org test-org",
			},
		},
		{
			name:    "Generate with org-list flag",
			command: "generate",
			flags: map[string]interface{}{
				"enterprise-slug":              "my-enterprise",
				"github-enterprise-server-url": "github.company.com",
				"org-list":                     "orgs.csv",
			},
			expected: []string{
				"gh security-config generate",
				"-e my-enterprise",
				"-u github.company.com",
				"-l orgs.csv",
			},
		},
		{
			name:    "Generate with concurrency",
			command: "generate",
			flags: map[string]interface{}{
				"enterprise-slug": "my-enterprise",
				"all-orgs":        true,
				"concurrency":     5,
			},
			expected: []string{
				"gh security-config generate",
				"-e my-enterprise",
				"--all-orgs",
				"-c 5",
			},
		},
		{
			name:    "Generate with delay",
			command: "generate",
			flags: map[string]interface{}{
				"enterprise-slug": "my-enterprise",
				"all-orgs":        true,
				"delay":           30,
			},
			expected: []string{
				"gh security-config generate",
				"-e my-enterprise",
				"--all-orgs",
				"-d 30",
			},
		},
		{
			name:    "Generate with force and copy-from-org",
			command: "generate",
			flags: map[string]interface{}{
				"enterprise-slug": "my-enterprise",
				"all-orgs":        true,
				"force":           true,
				"copy-from-org":   "source-org",
			},
			expected: []string{
				"gh security-config generate",
				"-e my-enterprise",
				"--all-orgs",
				"-f",
				"-o source-org",
			},
		},
		{
			name:    "Apply command",
			command: "apply",
			flags: map[string]interface{}{
				"enterprise-slug":              "my-enterprise",
				"github-enterprise-server-url": "github.company.com",
				"template-org":                 "template-org",
				"all-orgs":                     true,
			},
			expected: []string{
				"gh security-config apply",
				"-e my-enterprise",
				"-u github.company.com",
				"-t template-org",
				"--all-orgs",
			},
		},
		{
			name:    "Modify command",
			command: "modify",
			flags: map[string]interface{}{
				"enterprise-slug": "my-enterprise",
				"template-org":    "template-org",
				"org":             "test-org",
			},
			expected: []string{
				"gh security-config modify",
				"-e my-enterprise",
				"-t template-org",
				"--org test-org",
			},
		},
		{
			name:    "Delete command",
			command: "delete",
			flags: map[string]interface{}{
				"enterprise-slug": "my-enterprise",
				"template-org":    "template-org",
				"org-list":        "orgs.csv",
			},
			expected: []string{
				"gh security-config delete",
				"-e my-enterprise",
				"-t template-org",
				"-l orgs.csv",
			},
		},
		{
			name:    "String with spaces gets quoted",
			command: "generate",
			flags: map[string]interface{}{
				"enterprise-slug":              "my enterprise",
				"github-enterprise-server-url": "github.company.com",
				"all-orgs":                     true,
			},
			expected: []string{
				"gh security-config generate",
				"-e \"my enterprise\"",
				"-u github.company.com",
				"--all-orgs",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildReplicationCommand(tt.command, tt.flags)

			// Check that all expected substrings are present
			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("BuildReplicationCommand() result missing expected substring:\n  Expected: %s\n  Got: %s", expected, result)
				}
			}

			// Check that the command starts with the right prefix
			expectedPrefix := "gh security-config " + tt.command
			if !strings.HasPrefix(result, expectedPrefix) {
				t.Errorf("BuildReplicationCommand() result should start with %q, got %q", expectedPrefix, result)
			}
		})
	}
}

func TestQuoteIfNeeded(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No spaces - no quotes",
			input:    "test-org",
			expected: "test-org",
		},
		{
			name:     "With spaces - add quotes",
			input:    "test org",
			expected: "\"test org\"",
		},
		{
			name:     "Multiple spaces - add quotes",
			input:    "my enterprise name",
			expected: "\"my enterprise name\"",
		},
		{
			name:     "Empty string - no quotes",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := quoteIfNeeded(tt.input)
			if result != tt.expected {
				t.Errorf("quoteIfNeeded() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetShortFlag(t *testing.T) {
	tests := []struct {
		flagName string
		expected string
	}{
		{"org-list", "l"},
		{"concurrency", "c"},
		{"delay", "d"},
		{"enterprise-slug", "e"},
		{"github-enterprise-server-url", "u"},
		{"dependabot-alerts-available", "a"},
		{"dependabot-security-updates-available", "s"},
		{"copy-from-org", "o"},
		{"force", "f"},
		{"template-org", "t"},
		{"unknown-flag", ""}, // Should return empty string for unknown flags
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			result := getShortFlag(tt.flagName)
			if result != tt.expected {
				t.Errorf("getShortFlag(%q) = %q, want %q", tt.flagName, result, tt.expected)
			}
		})
	}
}
