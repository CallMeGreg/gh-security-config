# GitHub API Docs

## Create a code security configuration

Creates a code security configuration in an organization.

The authenticated user must be an administrator or security manager for the organization to use this endpoint.

OAuth app tokens and personal access tokens (classic) need the write:org scope to use this endpoint.

Note

Only installed security products may be specified in the request body. Specifying an uninstalled security product will result in a validation error.

Fine-grained access tokens for "Create a code security configuration"
This endpoint works with the following fine-grained token types:

GitHub App user access tokens
GitHub App installation access tokens
Fine-grained personal access tokens
The fine-grained token must have the following permission set:

"Administration" organization permissions (write)
Parameters for "Create a code security configuration"
Headers
Name, Type, Description
accept string
Setting to application/vnd.github+json is recommended.

Path parameters
Name, Type, Description
org string Required
The organization name. The name is not case sensitive.

Body parameters
Name, Type, Description
name string Required
The name of the code security configuration. Must be unique within the organization.

description string Required
A description of the code security configuration

advanced_security string
The enablement status of GitHub Advanced Security

Default: disabled

Can be one of: enabled, disabled

dependency_graph string
The enablement status of Dependency Graph. Dependency Graph is not configurable in GitHub Enterprise Server.

Default: enabled

Can be one of: enabled, disabled, not_set

dependabot_alerts string
The enablement status of Dependabot alerts

Default: disabled

Can be one of: enabled, disabled, not_set

dependabot_security_updates string
The enablement status of Dependabot security updates

Default: disabled

Can be one of: enabled, disabled, not_set

code_scanning_default_setup string
The enablement status of code scanning default setup

Default: disabled

Can be one of: enabled, disabled, not_set

secret_scanning string
The enablement status of secret scanning

Default: disabled

Can be one of: enabled, disabled, not_set

secret_scanning_push_protection string
The enablement status of secret scanning push protection

Default: disabled

Can be one of: enabled, disabled, not_set

secret_scanning_delegated_bypass string
The enablement status of secret scanning delegated bypass

Default: disabled

Can be one of: enabled, disabled, not_set

secret_scanning_delegated_bypass_options object
Feature options for secret scanning delegated bypass

Properties of secret_scanning_delegated_bypass_options
secret_scanning_validity_checks string
The enablement status of secret scanning validity checks

Default: disabled

Can be one of: enabled, disabled, not_set

secret_scanning_non_provider_patterns string
The enablement status of secret scanning non provider patterns

Default: disabled

Can be one of: enabled, disabled, not_set

private_vulnerability_reporting string
The enablement status of private vulnerability reporting

Default: disabled

Can be one of: enabled, disabled, not_set

enforcement string
The enforcement status for a security configuration

Default: enforced

Can be one of: enforced, unenforced

HTTP response status codes for "Create a code security configuration"
Status code	Description
201	
Successfully created code security configuration

## GitHub CLI example

gh api \
  --method POST \
  -H "Accept: application/vnd.github+json" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  --hostname HOSTNAME \
  /orgs/ORG/code-security/configurations \
   -f 'name=octo-org recommended settings' -f 'description=This is a code security configuration for octo-org' -f 'advanced_security=enabled' -f 'dependabot_alerts=enabled' -f 'dependabot_security_updates=not_set' -f 'secret_scanning=enabled'