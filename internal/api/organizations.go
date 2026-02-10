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

// GetOrganizations returns organization list from one of three sources:
// 1) A single org name (--org)
// 2) A CSV file of org names (--org-list)
// 3) All orgs in the enterprise (--all-orgs)
func GetOrganizations(enterprise, org, orgListPath string, allOrgs bool) ([]string, error) {
	if org != "" {
		pterm.Info.Printf("Targeting single organization: %s\n", pterm.Green(org))
		pterm.Println()
		return []string{org}, nil
	}

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

		// Show the list of organizations that will be targeted
		if len(csvOrgs) <= 10 {
			pterm.Info.Println("Organizations to be targeted:")
			for _, org := range csvOrgs {
				pterm.Printf("  - %s\n", pterm.Green(org))
			}
		} else {
			pterm.Info.Printf("Organizations to be targeted: %d organizations\n", len(csvOrgs))
		}
		pterm.Println()

		return csvOrgs, nil
	}

	if allOrgs {
		// Use existing enterprise API fetching
		pterm.Info.Println("Fetching all organizations from enterprise...")
		orgs, err := FetchOrganizations(enterprise)
		if err != nil {
			return nil, err
		}
		pterm.Success.Printf("Found %d organizations in enterprise '%s'\n", len(orgs), enterprise)
		return orgs, nil
	}

	return nil, fmt.Errorf("one of --org, --org-list, or --all-orgs must be specified")
}

// formatCursor formats the cursor for GraphQL pagination
func formatCursor(cursor *string) string {
	if cursor == nil {
		return "null"
	}
	return fmt.Sprintf(`"%s"`, *cursor)
}
