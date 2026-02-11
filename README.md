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
- **`--github-enterprise-server-url string`** (`-u`) - GitHub Enterprise Server URL (e.g., github.company.com)
- **`--dependabot-alerts-available string`** (`-a`) - Whether Dependabot Alerts are available in your GHES instance (true/false)
- **`--dependabot-security-updates-available string`** (`-s`) - Whether Dependabot Security Updates are available in your GHES instance (true/false)

### Generate Command Flags

The `generate` command has additional flags:

- **`--copy-from-org string`** (`-o`) - Organization name to copy an existing configuration from
- **`--force`** (`-f`) - Force deletion of existing configurations with the same name before creating new ones

### Apply, Delete, and Modify Command Flags

The `apply`, `delete`, and `modify` commands have an additional required flag:

- **`--template-org string`** (`-t`) - Template organization to fetch security configurations from. This organization is used as the source of truth for security configuration settings. If not provided, you will be prompted to enter it interactively.

### Basic Usage Examples

```bash
# Create a new security configuration
gh security-config generate

# Apply an existing security configuration to all orgs
gh security-config apply --all-orgs

# Modify a security configuration
gh security-config modify

# Delete a security configuration from a CSV list of organizations
gh security-config delete --org-list orgs.csv

# Use flags to skip interactive prompts
gh security-config generate --all-orgs -e my-enterprise -u github.mycompany.com -a true -s false

# Use concurrent processing for faster execution (up to 20 organizations at once)
gh security-config generate --all-orgs --concurrency 5

# Use delayed processing to avoid rate limits and reduce system overhead (30 second delay between organizations)
gh security-config generate --org-list orgs.csv --delay 30
```

### Copying Security Configurations

The `--copy-from-org` flag allows you to copy an existing security configuration from one organization and apply it to other organizations in your enterprise. This is useful for:

- **Standardizing configurations**: Copy a well-tested configuration across multiple organizations
- **Quick setup**: Avoid recreating similar configurations from scratch
- **Configuration migration**: Move configurations between organizations

#### How it works:

1. **Source Organization Access**: You must be an owner of the source organization to copy configurations
2. **Configuration Selection**: Choose from available security configurations in the source organization
3. **Settings Review**: Review the configuration details that will be copied
4. **Target Filtering**: The source organization is automatically excluded from target organizations to prevent self-copying

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


#### Sequential Processing with Delay (`--delay`)

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

### Apply Security Configurations

The extension will guide you through:

1. **Enterprise Setup**: Enter your GitHub Enterprise slug and server URL (if using GitHub Enterprise Server)
2. **Template Organization**: Specify the template organization from which to fetch the security configuration
3. **Organization Targeting**: If not provided via flags, select whether to target all orgs, a single org, or orgs from a CSV file
4. **Configuration Selection**: Select from available security configurations in the template organization
5. **Repository Selection**: Choose which repositories should have the configuration applied
6. **Confirmation**: Review the operation summary and confirm application

> [!NOTE]
> The apply operation fetches the security configuration from the specified template organization and applies it to repositories across the targeted organizations. The template org serves as the source of truth for the configuration settings.

### Delete Security Configurations

The extension will guide you through:

1. **Enterprise Setup**: Enter your GitHub Enterprise slug and server URL (if using GitHub Enterprise Server)
2. **Template Organization**: Specify the template organization from which to reference the security configuration
3. **Organization Targeting**: If not provided via flags, select whether to target all orgs, a single org, or orgs from a CSV file
4. **Configuration Selection**: Specify the name of the security configuration to delete
5. **Confirmation**: Review the operation summary and confirm deletion (defaults to cancel for safety)

> [!WARNING]
> The delete operation will remove the specified security configuration from the targeted organizations. This action cannot be undone. Repositories will retain their security settings but will no longer be associated with the configuration.

### Modify Security Configurations

The extension will guide you through:

1. **Enterprise Setup**: Enter your GitHub Enterprise slug and server URL (if using GitHub Enterprise Server)
2. **Template Organization**: Specify the template organization from which to fetch the current security configuration
3. **Organization Targeting**: If not provided via flags, select whether to target all orgs, a single org, or orgs from a CSV file
4. **Configuration Selection**: Specify the name of the security configuration to modify
5. **Current Settings Display**: View the current configuration settings and description from the template organization
4. **Name Update**: Update the configuration name (optional)
5. **Description Update**: Update the configuration description (optional)
6. **Settings Update**: Interactively update each security setting with options to keep current values
7. **Confirmation**: Review the changes and confirm modification before execution

> [!NOTE]
> The modify operation will update the specified security configuration across ALL organizations in the enterprise where it exists. Organizations without the configuration will be skipped.

## Security Settings

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
3. Submit a pull request

## License

This tool is licensed under the MIT License.
