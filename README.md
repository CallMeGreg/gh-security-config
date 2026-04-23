# GitHub Security Configuration CLI Extension

A GitHub CLI extension to create, apply, modify, or delete security configurations across many organizations in a GitHub Enterprise.

> [!NOTE]
> For GHES 3.16+ and GitHub Enterprise Cloud (GHEC) it's recommended to use a single [Enterprise Security Configuration](https://docs.github.com/en/enterprise-cloud@latest/admin/managing-code-security/securing-your-enterprise/about-security-configurations) instead of creating security configurations per organization.

## Pre-requisites

1. [GitHub CLI](https://github.com/cli/cli#installation)
2. [GitHub Advanced Security](https://docs.github.com/en/enterprise-cloud@latest/get-started/learning-about-github/about-github-advanced-security) licenses and availability in your organizations.
3. Confirm that you are authenticated with an account that has access to the enterprise and organizations you would like to interact with. You can check your authentication status by running:

```
gh auth status
```

Ensure that you have the necessary scopes (`read:enterprise` and `admin:org`). You can add scopes by running:

```
gh auth login -s "read:enterprise,admin:org"
```

> [!IMPORTANT]
> Enterprise admins do not inherently have access to all of the organizations in the enterprise. You must ensure that your account has the necessary permissions to access the organizations you want to modify. To elevate your permissions for an organization, refer to these [GitHub docs](https://docs.github.com/en/enterprise-server@3.15/admin/managing-accounts-and-repositories/managing-organizations-in-your-enterprise/managing-your-role-in-an-organization-owned-by-your-enterprise).

## Installation

To install this extension, run the following command:
```
gh extension install CallMeGreg/gh-security-config
```

## Usage

The extension provides four main commands for managing security configurations across enterprise organizations:

### Commands

- **`generate`** - Create and optionally apply new security configurations across organizations
- **`apply`** - Apply existing security configurations to repositories across organizations
- **`modify`** - Update existing security configurations across organizations
- **`delete`** - Remove existing security configurations from organizations

### Quick Start

```bash
# Create a new security configuration
gh security-config generate

# Apply an existing security configuration to all orgs
gh security-config apply --all-orgs

# Modify a security configuration for a CSV list of organizations
gh security-config modify --org-list orgs.csv

# Delete a security configuration from a single org
gh security-config delete --org first-org

# Use flags to skip interactive prompts
gh security-config generate --all-orgs -e my-enterprise -u github.mycompany.com -a true -s false

# Use concurrent processing for faster execution (up to 20 organizations at once)
gh security-config generate --all-orgs --concurrency 5

# Use delayed processing to avoid rate limits and reduce system overhead (30 second delay between organizations)
gh security-config generate --org-list orgs.csv --delay 30

# Run generate non-interactively in a single call
gh security-config generate \
  --all-orgs -e my-enterprise -u github.mycompany.com -a true -s false \
  --config-name "org-default" --config-description "Org default security configuration" \
  --advanced-security enabled \
  --dependabot-alerts enabled --dependabot-security-updates not_set \
  --secret-scanning enabled --secret-scanning-push-protection enabled \
  --secret-scanning-non-provider-patterns not_set \
  --enforcement enforced \
  --scope all --set-as-default true --skip-confirmation-message true
```

> [!TIP]
> The replication command printed at the end of each run emits the full set of flags used (including any values chosen interactively), so the exact invocation can be re-run non-interactively.

### Persistent Flags

These flags are available on all commands:

#### Organization Targeting (mutually exclusive)

- **`--org string`** - Target a single organization by name
- **`--org-list string`** (`-l`) - Path to CSV file containing organization names to target (one per line, no header)
- **`--all-orgs`** - Target all organizations in the enterprise

#### Other Flags

- **`--concurrency int`** (`-c`) - Number of concurrent requests (1-20, default: 1, mutually exclusive with `--delay`)
- **`--delay int`** (`-d`) - Delay in seconds between organizations (1-600, mutually exclusive with `--concurrency`)
- **`--enterprise-slug string`** (`-e`) - GitHub Enterprise slug (e.g., github)
- **`--github-enterprise-server-url string`** (`-u`) - GitHub Enterprise URL (e.g., github.company.com)
- **`--dependabot-alerts-available string`** (`-a`) - Whether Dependabot Alerts are available in your GHES instance (true/false)
- **`--dependabot-security-updates-available string`** (`-s`) - Whether Dependabot Security Updates are available in your GHES instance (true/false)
- **`--config-name string`** (`-n`) - Name of the security configuration to operate on. Replaces the interactive configuration-name prompt for each command (the meaning is command-specific: the name to create in `generate`, the name to select in `apply`/`delete`/`modify`, or the name of the source config in `generate --copy-from-org`).
- **`--skip-confirmation-message string`** - Automatically approve the final confirmation prompt for any command (`true`/`false`).
- **`--log-level string`** - Minimum log level for output (`info`, `warning`, `error`; default: `warning`). When set to `info`, a success message is printed for each organization that is processed successfully.

#### `generate` Command Flags

| Flag | Interactive prompt it replaces |
|------|--------------------------------|
| `--config-description` | "Enter security configuration description" |
| `--advanced-security` | "GitHub Advanced Security" (`enabled`, `disabled`) |
| `--dependabot-alerts` | "Dependabot Alerts" (`enabled`, `disabled`, `not_set`) |
| `--dependabot-security-updates` | "Dependabot Security Updates" (`enabled`, `disabled`, `not_set`) |
| `--secret-scanning` | "Secret Scanning" (`enabled`, `disabled`, `not_set`) |
| `--secret-scanning-push-protection` | "Secret Scanning Push Protection" (`enabled`, `disabled`, `not_set`) |
| `--secret-scanning-non-provider-patterns` | "Secret Scanning Non-Provider Patterns" (`enabled`, `disabled`, `not_set`) |
| `--enforcement` | "Enforcement Status" (`enforced`, `unenforced`) |
| `--scope` | "Select repositories to attach configuration to" (`all`, `public`, `private_or_internal`, `none`) |
| `--set-as-default` | "Set this configuration as default for new repositories?" (`true`, `false`) |
| `--overwrite` | Overwrite any existing configuration with the same name instead of skipping (`true`, `false`) |

#### `apply` Command Flags

| Flag | Interactive prompt it replaces |
|------|--------------------------------|
| `--config-source` | Disambiguates `--config-name` when the same name exists at both levels (`organization`, `enterprise`) |
| `--scope` | "Select repositories to attach configuration to" (`all`, `public`, `private_or_internal`) |
| `--set-as-default` | "Set this configuration as default for new repositories?" (`true`, `false`) |

#### `delete` Command Flags

The `delete` command uses only the universal `--config-name` and `--skip-confirmation-message` flags (plus `--template-org`). No additional command-specific input flags.

#### `modify` Command Flags

| Flag | Interactive prompt it replaces |
|------|--------------------------------|
| `--new-name` | "Enter updated security configuration name" (omit to keep the current name) |
| `--new-description` | "Enter updated security configuration description" (omit to keep the current description) |
| `--advanced-security` | Update prompt for GitHub Advanced Security (`enabled`, `disabled`) |
| `--dependabot-alerts` | Update prompt for Dependabot Alerts (`enabled`, `disabled`, `not_set`) |
| `--dependabot-security-updates` | Update prompt for Dependabot Security Updates (`enabled`, `disabled`, `not_set`) |
| `--secret-scanning` | Update prompt for Secret Scanning (`enabled`, `disabled`, `not_set`) |
| `--secret-scanning-push-protection` | Update prompt for Secret Scanning Push Protection (`enabled`, `disabled`, `not_set`) |
| `--secret-scanning-non-provider-patterns` | Update prompt for Secret Scanning Non-Provider Patterns (`enabled`, `disabled`, `not_set`) |
| `--enforcement` | Update prompt for Enforcement Status (`enforced`, `unenforced`) |

> [!NOTE]
> When using `--copy-from-org`, you can still customize the repository attachment scope and default setting for the target organizations, even though the security settings themselves are copied from the source.

### Concurrency and Performance

All commands support two execution modes for processing multiple organizations:

#### Concurrent Processing (`--concurrency`)

Process multiple organizations simultaneously to improve performance:

- **Default**: `1` (sequential processing, maintains existing behavior)
- **Range**: `1-20` (validated to prevent excessive API usage)
- **Usage**: Available on all commands (`generate`, `apply`, `modify`, `delete`)
- **Benefits**: Significantly reduces total processing time for large numbers of organizations

> [!WARNING]
> **Rate Limiting Considerations**: Setting concurrency higher than 1 increases the likelihood of encountering GitHub's secondary rate limits. To avoid rate limiting issues, consider [exempting the user from rate limits](https://docs.github.com/en/enterprise-server@3.15/admin/administering-your-instance/administering-your-instance-from-the-command-line/command-line-utilities#ghe-config).


#### Sequential Processing with Optional Delay (`--delay`)

Process organizations one at a time with a configurable delay between each:

- **Range**: `1-600` seconds (validated to prevent unreasonable delays)
- **Usage**: Available on all commands (`generate`, `apply`, `modify`, `delete`)
- **Benefits**: Helps avoid rate limiting issues and provides controlled processing pace

### Error Handling and Requirements

#### Dependabot Feature Availability

Dependabot Alerts and Security Updates have different availability requirements:

- **Dependabot Alerts**: Available when GitHub Connect, Dependency Graph, and Dependabot are enabled
- **Dependabot Security Updates**: Available when Dependabot Alerts and GitHub Actions are enabled

**Checking Availability**: Navigate to `Enterprise settings` → `Settings` → `Code security and analysis` to verify which features are available.

## Security Configuration Settings

The extension allows you to set the following features within the security configuration:

| Setting | Description | Options |
|---------|-------------|---------|
| GitHub Advanced Security | The enablement status of GitHub Advanced Security | `enabled`, `disabled` |
| Dependabot Alerts | Detect vulnerable dependencies | `enabled`, `disabled`, `not_set` |
| Dependabot Security Updates | Automatically create pull requests to update vulnerable dependencies | `enabled`, `disabled`, `not_set` |
| Secret Scanning | Detect secrets in code | `enabled`, `disabled`, `not_set` |
| Secret Scanning Push Protection | Block commits with secrets | `enabled`, `disabled`, `not_set` |
| Secret Scanning Non-Provider Patterns | Scan for [non-provider patterns](https://docs.github.com/en/enterprise-cloud@latest/code-security/secret-scanning/using-advanced-secret-scanning-and-push-protection-features/non-provider-patterns) | `enabled`, `disabled`, `not_set` |
| Enforcement | Restrict setting changes at the repository level | `enforced`, `unenforced` |

## Repository Attachment Scopes

When attaching configurations to repositories, you can choose:

- **all**: Apply to all repositories in the organization
- **public**: Apply only to public repositories
- **private_or_internal**: Apply only to private and internal repositories
- **none**: Create the configuration without applying it to any repositories

## Demos

### Create and apply a new organization security configuration in every org

![Generate demo](demo/generate.gif)

### Apply an existing organization configuration to all repos across every org

![Apply demo](demo/apply.gif)

## Development

To build the extension locally:

```bash
go build -o gh-security-config
```

To run the extension locally:

```bash
./gh-security-config --help
```

## Contributing

1. Fork the repository
2. Make your changes
3. Open a pull request

## License

This tool is licensed under the MIT License.
