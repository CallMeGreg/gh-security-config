# GitHub API Docs

## Attach a configuration to repositories

Attach a code security configuration to a set of repositories. If the repositories specified are already attached to a configuration, they will be re-attached to the provided configuration.

If insufficient GHAS licenses are available to attach the configuration to a repository, only free features will be enabled.

The authenticated user must be an administrator or security manager for the organization to use this endpoint.

OAuth app tokens and personal access tokens (classic) need the write:org scope to use this endpoint.

Fine-grained access tokens for "Attach a configuration to repositories"
This endpoint works with the following fine-grained token types:

GitHub App user access tokens
GitHub App installation access tokens
Fine-grained personal access tokens
The fine-grained token must have the following permission set:

"Administration" organization permissions (write)
Parameters for "Attach a configuration to repositories"
Headers
Name, Type, Description
accept string
Setting to application/vnd.github+json is recommended.

Path parameters
Name, Type, Description
org string Required
The organization name. The name is not case sensitive.

configuration_id integer Required
The unique identifier of the code security configuration.

Body parameters
Name, Type, Description
scope string Required
The type of repositories to attach the configuration to. selected means the configuration will be attached to only the repositories specified by selected_repository_ids

Can be one of: all, all_without_configurations, public, private_or_internal, selected

selected_repository_ids array of integers
An array of repository IDs to attach the configuration to. You can only provide a list of repository ids when the scope is set to selected.

HTTP response status codes for "Attach a configuration to repositories"
Status code	Description
202	
Accepted

## GitHub CLI example

gh api \
  --method POST \
  -H "Accept: application/vnd.github+json" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  --hostname HOSTNAME \
  /orgs/ORG/code-security/configurations/CONFIGURATION_ID/attach \
  --input - <<< '{
  "scope": "selected",
  "selected_repository_ids": [
    32,
    91
  ]
}'
