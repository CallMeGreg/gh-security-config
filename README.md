# GitHub Security Configuration CLI Extension

A GitHub CLI extension to create and apply security configurations across many organizations in a GitHub Enterprise.

> [!NOTE]
> This extension is intended for GHES 3.15 and currently only supports configuring GitHub Advanced Security and Secret Scanning features as part of a security configuration. For GitHub Enterprise Server 3.16+ and GitHub Enterprise Cloud it's recommended to use [Enterprise Security Configurations](https://docs.github.com/en/enterprise-cloud@latest/admin/managing-code-security/securing-your-enterprise/about-security-configurations) instead of this solution.

## Pre-requisites

1. [GitHub Advanced Security](https://docs.github.com/en/enterprise-cloud@latest/get-started/learning-about-github/about-github-advanced-security) licenses and availability in your organizations.
2. [GitHub CLI](https://github.com/cli/cli#installation)
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

The extension provides three main commands for managing security configurations across enterprise organizations:

### Commands

- **`generate`** - Create and apply new security configurations across organizations
- **`delete`** - Remove existing security configurations from organizations  
- **`modify`** - Update existing security configurations across organizations

### Persistent Flags

These flags are available on all commands:

- **`--org-list string`** - Path to CSV file containing organization names to target (one per line, no header)
- **`--concurrency int`** - Number of concurrent requests (1-20, default: 1)

### Generate Command Flags

The `generate` command has additional flags:

- **`--copy-from-org string`** - Organization name to copy an existing configuration from
- **`--force`** - Force deletion of existing configurations with the same name before creating new ones

### Basic Usage Examples

```bash
# Create a new security configuration interactively
gh security-config generate

# Delete a security configuration from specific organizations
gh security-config delete --org-list organizations.csv

# Modify configurations with concurrent processing
gh security-config modify --concurrency 5
```

### Organization Targeting

By default, all commands target every organization in the specified enterprise. You can limit the scope using the `--org-list` flag:

- **CSV Format**: Create a CSV file with one organization name per line (no header row required)
- **Example CSV**: See [example-organizations.csv](example-organizations.csv) for the correct format
- **Error Handling**: If an organization from the CSV is not found or accessible, the tool will show a warning and continue with other organizations

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

### Interactive Workflow

All commands provide an interactive workflow that guides you through:

1. **Enterprise Setup**: Enter your GitHub Enterprise slug and server URL (if using GitHub Enterprise Server)
2. **Configuration Details**: Define or select security configurations based on the command:
   - **Generate**: Create new configurations or copy from existing organizations
   - **Delete**: Specify configuration name to remove
   - **Modify**: Select configuration and update settings
3. **Repository Scope**: Choose which repositories to apply configurations to (generate only):
   - `all` - All repositories
   - `public` - Public repositories only
   - `private_or_internal` - Private and internal repositories only
4. **Default Setting**: Optionally set configurations as default for new repositories (generate only)
5. **Confirmation**: Review and confirm the operation before execution

### Concurrency and Performance

All commands support concurrent requests using the `--concurrency` flag to improve performance when working with many organizations:

```bash
# Process organizations sequentially (default)
gh security-config generate

# Process up to 5 organizations concurrently
gh security-config generate --concurrency 5

# Use maximum concurrency for fastest processing
gh security-config delete --concurrency 20 --org-list organizations.csv
```

#### Concurrency Settings

- **Default**: `1` (sequential processing, maintains existing behavior)
- **Range**: `1-20` (validated to prevent excessive API usage)
- **Usage**: Available on all commands (`generate`, `delete`, `modify`)

#### Performance Benefits

- **Faster Execution**: Significantly reduces total processing time for large numbers of organizations
- **Configurable**: Choose concurrency level based on your needs and environment
- **Progress Tracking**: Real-time progress updates work seamlessly with concurrent processing

> [!WARNING]
> **Rate Limiting Considerations**: Setting concurrency higher than 1 increases the likelihood of encountering GitHub's secondary rate limits. To avoid rate limiting issues, consider [exempting the user from rate limits](https://docs.github.com/en/enterprise-server@3.15/admin/administering-your-instance/administering-your-instance-from-the-command-line/command-line-utilities#ghe-config).

### Delete Security Configurations

Run the interactive security configuration deletion:

```bash
gh security-config delete
```

To target specific organizations using a CSV file:

```bash
gh security-config delete --org-list path/to/organizations.csv
```

For faster processing with many organizations, use concurrency:

```bash
gh security-config delete --org-list path/to/organizations.csv --concurrency 3
```

The extension will guide you through:

1. **Enterprise Setup**: Enter your GitHub Enterprise slug and server URL (if using GitHub Enterprise Server)
2. **Configuration Selection**: Specify the name of the security configuration to delete
3. **Confirmation**: Review the operation summary and confirm deletion (defaults to cancel for safety)

> [!WARNING]
> The delete operation will remove the specified security configuration from ALL organizations in the enterprise. This action cannot be undone. Repositories will retain their security settings but will no longer be associated with the configuration.

### Modify Security Configurations

Run the interactive security configuration modification:

```bash
gh security-config modify
```

To target specific organizations using a CSV file:

```bash
gh security-config modify --org-list path/to/organizations.csv
```

For faster processing with many organizations, use concurrency:

```bash
gh security-config modify --org-list path/to/organizations.csv --concurrency 5
```

The extension will guide you through:

1. **Enterprise Setup**: Enter your GitHub Enterprise slug and server URL (if using GitHub Enterprise Server)
2. **Configuration Selection**: Specify the name of the security configuration to modify
3. **Current Settings Display**: View the current configuration settings and description
4. **Settings Update**: Interactively update each security setting with options to keep current values
5. **Confirmation**: Review the changes and confirm modification before execution

> [!NOTE]
> The modify operation will update the specified security configuration across ALL organizations in the enterprise where it exists. Organizations without the configuration will be skipped.

## Features

- 🏢 **Enterprise-wide Management**: Automatically discovers and processes all organizations in your enterprise
- ⚡ **Concurrent Processing**: Configurable concurrency (1-20) for faster processing of many organizations
- 🔒 **Comprehensive Security Settings**: Configure GitHub Advanced Security features and related settings:
  - GitHub Advanced Security
  - Secret Scanning
  - Secret Scanning Push Protection
  - Secret Scanning Non-Provider Patterns
  - Enforcement
- 🎯 **Flexible Targeting**: Choose which repositories to apply configurations to
- ➕ **Configuration Generation**: Create and apply security configurations across all enterprise organizations
- 📋 **Configuration Copying**: Copy existing security configurations from one organization to others in the enterprise
- ✏️ **Configuration Modification**: Update existing security configurations across all enterprise organizations with selective setting changes
- ❌ **Configuration Deletion**: Safely delete security configurations from all enterprise organizations with confirmation prompts
- ⚙️ **Default Configuration**: Optionally set configurations as defaults for new repositories
- 📊 **Progress Tracking**: Visual progress indicators with concurrent operation support

## Security Settings

The extension allows you to set the following features within the security configuration:

| Setting | Description | Options |
|---------|-------------|---------|
| GitHub Advanced Security | The enablement status of GitHub Advanced Security | `enabled`, `disabled` |
| Secret Scanning | Detect secrets in code | `enabled`, `disabled`, `not_set` |
| Secret Scanning Push Protection | Block commits with secrets | `enabled`, `disabled`, `not_set` |
| Secret Scanning Non-Provider Patterns | Scan for [non-provider patterns](https://docs.github.com/en/enterprise-cloud@latest/code-security/secret-scanning/using-advanced-secret-scanning-and-push-protection-features/non-provider-patterns) | `enabled`, `disabled`, `not_set` |
| Enforcement | Restrict setting changes at the repository level | `enforced`, `unenforced` |

## Repository Attachment Scopes

When attaching configurations to repositories, you can choose:

- **all**: Apply to all repositories in the organization
- **public**: Apply only to public repositories
- **private_or_internal**: Apply only to private and internal repositories

## Example

![Demo of gh-security-config generate](docs/gh-security-config-demo.gif)

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
