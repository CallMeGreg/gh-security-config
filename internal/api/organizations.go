package api

import (
	"encoding/json"
	"fmt"

	"github.com/cli/go-gh/v2"
	"github.com/pterm/pterm"

	"github.com/callmegreg/gh-security-config/internal/utils"
)

// FetchOrganizations fetches all organizations from an enterprise using GraphQL
func FetchOrganizations(enterprise string) ([]string, error) {
	const maxPerPage = 100
	var orgs []string
	var cursor *string

	for {
		query := fmt.Sprintf(`{
			enterprise(slug: "%s") {
				organizations(first: %d, after: %s) {
					nodes {
						login
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}`, enterprise, maxPerPage, formatCursor(cursor))

		response, stderr, err := gh.Exec("api", "graphql", "-f", "query="+query)
		if err != nil {
			pterm.Error.Printf("Failed to fetch organizations for enterprise '%s': %v\n", enterprise, err)
			pterm.Error.Printf("GraphQL query: %s\n", query)
			pterm.Error.Printf("gh CLI stderr: %s\n", stderr.String())
			return nil, err
		}

		var result struct {
			Data struct {
				Enterprise struct {
					Organizations struct {
						Nodes []struct {
							Login string `json:"login"`
						}
						PageInfo struct {
							HasNextPage bool   `json:"hasNextPage"`
							EndCursor   string `json:"endCursor"`
						} `json:"pageInfo"`
					} `json:"organizations"`
				} `json:"enterprise"`
			} `json:"data"`
		}

		if err := json.Unmarshal(response.Bytes(), &result); err != nil {
			pterm.Error.Printf("Failed to parse organizations data for enterprise '%s': %v\n", enterprise, err)
			return nil, err
		}

		for _, org := range result.Data.Enterprise.Organizations.Nodes {
			orgs = append(orgs, org.Login)
		}

		if !result.Data.Enterprise.Organizations.PageInfo.HasNextPage {
			break
		}
		cursor = &result.Data.Enterprise.Organizations.PageInfo.EndCursor
	}

	return orgs, nil
}

// GetOrganizations returns organization list either from CSV file or from enterprise API
func GetOrganizations(enterprise, orgListPath string) ([]string, error) {
	if orgListPath != "" {
		pterm.Info.Printf("Reading organizations from CSV file: %s\n", orgListPath)
		csvOrgs, err := utils.ReadOrganizationsFromCSV(orgListPath)
		if err != nil {
			return nil, err
		}
		if len(csvOrgs) == 0 {
			return nil, fmt.Errorf("no valid organizations found in CSV file")
		}
		pterm.Success.Printf("Found %d organizations in CSV file\n", len(csvOrgs))

		// Fetch all organizations from enterprise to validate against
		pterm.Info.Println("Fetching organizations from enterprise to validate CSV list...")
		enterpriseOrgs, err := FetchOrganizations(enterprise)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch enterprise organizations for validation: %w", err)
		}
		pterm.Success.Printf("Found %d organizations in enterprise '%s'\n", len(enterpriseOrgs), enterprise)

		// Create a map for faster lookup
		enterpriseOrgMap := make(map[string]bool)
		for _, org := range enterpriseOrgs {
			enterpriseOrgMap[org] = true
		}

		// Validate CSV organizations against enterprise organizations
		var validOrgs []string
		var invalidOrgs []string

		for _, org := range csvOrgs {
			if enterpriseOrgMap[org] {
				validOrgs = append(validOrgs, org)
			} else {
				invalidOrgs = append(invalidOrgs, org)
			}
		}

		// Warn about invalid organizations
		if len(invalidOrgs) > 0 {
			pterm.Warning.Printf("Found %d organizations in CSV that do not exist in enterprise '%s':\n", len(invalidOrgs), enterprise)
			for _, org := range invalidOrgs {
				pterm.Printf("  - %s (not found in enterprise)\n", pterm.Red(org))
			}
			pterm.Println()
		}

		// Check if we have any valid organizations left
		if len(validOrgs) == 0 {
			return nil, fmt.Errorf("no valid organizations found in CSV file that exist in enterprise '%s'", enterprise)
		}

		if len(invalidOrgs) > 0 {
			pterm.Info.Printf("Proceeding with %d valid organizations (skipping %d invalid)\n", len(validOrgs), len(invalidOrgs))
		}

		// Show the list of valid organizations that will be targeted
		pterm.Info.Println("Valid organizations to be targeted:")
		for _, org := range validOrgs {
			pterm.Printf("  - %s\n", pterm.Green(org))
		}
		pterm.Println()

		return validOrgs, nil
	}

	// Use existing enterprise API fetching
	pterm.Info.Println("Fetching organizations from enterprise...")
	orgs, err := FetchOrganizations(enterprise)
	if err != nil {
		return nil, err
	}
	pterm.Success.Printf("Found %d organizations in enterprise '%s'\n", len(orgs), enterprise)
	return orgs, nil
}

// formatCursor formats the cursor for GraphQL pagination
func formatCursor(cursor *string) string {
	if cursor == nil {
		return "null"
	}
	return fmt.Sprintf(`"%s"`, *cursor)
}
