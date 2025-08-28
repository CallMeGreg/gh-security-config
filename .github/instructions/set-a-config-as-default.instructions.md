# GitHub API Docs

## Set a code security configuration as a default for an organization

Sets a code security configuration as a default to be applied to new repositories in your organization.

This configuration will be applied to the matching repository type (all, none, public, private and internal) by default when they are created.

The authenticated user must be an administrator or security manager for the organization to use this endpoint.

OAuth app tokens and personal access tokens (classic) need the write:org scope to use this endpoint.

Note

The enablement status will only be returned for installed security products.

Fine-grained access tokens for "Set a code security configuration as a default for an organization"
This endpoint works with the following fine-grained token types:

GitHub App user access tokens
GitHub App installation access tokens
Fine-grained personal access tokens
The fine-grained token must have the following permission set:

"Administration" organization permissions (write)
Parameters for "Set a code security configuration as a default for an organization"
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
default_for_new_repos string
Specify which types of repository this security configuration should be applied to by default.

Can be one of: all, none, private_and_internal, public

HTTP response status codes for "Set a code security configuration as a default for an organization"
Status code	Description
200	
Default successfully changed.

403	
Forbidden

404	
Resource not found

## GitHub CLI example

gh api \
  --method PUT \
  -H "Accept: application/vnd.github+json" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  --hostname HOSTNAME \
  /orgs/ORG/code-security/configurations/CONFIGURATION_ID/defaults \
   -f 'default_for_new_repos=all'