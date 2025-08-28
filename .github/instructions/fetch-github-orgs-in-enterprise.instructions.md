# Fetch GitHub Organizations in an Enterprise

The following example shows how to fetch all organizations in an enterprise account in Go, using the GitHub GraphQL API.

```go
// FetchOrganizations fetches organizations for a given enterprise using the GitHub GraphQL API.
func FetchOrganizations(enterprise string, orgLimit int) ([]string, error) {
	if enterprise == "" {
		return nil, fmt.Errorf("--enterprise flag is required")
	}

	const maxPerPage = 100
	var orgs []string
	var cursor *string
	fetched := 0

	for {
		remaining := orgLimit - fetched
		if remaining > maxPerPage {
			remaining = maxPerPage
		}

		query := `{
			enterprise(slug: "` + enterprise + `") {
				organizations(first: ` + fmt.Sprintf("%d", remaining) + `, after: ` + formatCursor(cursor) + `) {
					nodes {
						login
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}`

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
			fetched++
			if fetched >= orgLimit {
				return orgs, nil
			}
		}

		if !result.Data.Enterprise.Organizations.PageInfo.HasNextPage {
			break
		}
		cursor = &result.Data.Enterprise.Organizations.PageInfo.EndCursor
	}

	return orgs, nil
}
```