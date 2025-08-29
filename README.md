# GitHub Security Configuration CLI Extension

A GitHub CLI extension to create and apply security configurations across all organizations in a GitHub Enterprise.

> [!NOTE]
> This extension currently only supports configuring GitHub Advanced Security and Secret Scanning features as part of a security configuration.

## Pre-requisites

1. For [GitHub Advanced Security](https://docs.github.com/en/enterprise-cloud@latest/get-started/learning-about-github/about-github-advanced-security) features, your organizations need appropriate licensing.
2. Install the GitHub CLI: https://github.com/cli/cli#installation
3. Confirm that you are authenticated with an account that has access to the enterprise and organizations you would like to interact with. You can check your authentication status by running:

```
gh auth status
```

Ensure that you have the necessary scopes (`read:enterprise` and `admin:org`). You can add scopes by running:

```
gh auth login -s "read:enterprise,admin:org"
```

> [!IMPORTANT]
> Enterprise admins do not inherently have access to all of the organizations in the enterprise. You must ensure that your account has the necessary permissions to access the organizations you want to modify.

## Installation

To install this extension, run the following command:
```
gh extension install CallMeGreg/gh-security-config
```

## Usage

Run the interactive security configuration generator:

```bash
gh security-config generate
```

The extension will guide you through:

1. **Enterprise Setup**: Enter your GitHub Enterprise slug and server URL (if using GitHub Enterprise Server)
2. **Security Configuration**: Define the name, description, and security settings for your configuration
3. **Repository Scope**: Choose which repositories to apply the configuration to:
   - `all` - All repositories
   - `public` - Public repositories only
   - `private_or_internal` - Private and internal repositories only
4. **Default Setting**: Optionally set the configuration as default for new repositories
5. **Confirmation**: Review and confirm the operation before execution

## Features

- 🏢 **Enterprise-wide Management**: Automatically discovers and processes all organizations in your enterprise
- 🔒 **Comprehensive Security Settings**: Configure GitHub Advanced Security features and related settings:
  - GitHub Advanced Security
  - Secret Scanning
  - Secret Scanning Push Protection
  - Secret Scanning Non-Provider Patterns
  - Enforcement
- 🎯 **Flexible Targeting**: Choose which repositories to apply configurations to
- ⚙️ **Default Configuration**: Optionally set configurations as defaults for new repositories
- 📊 **Progress Tracking**: Visual progress indicators
- 🖥️ **GitHub Enterprise Server Support**: Works with both GitHub.com and GitHub Enterprise Server

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

![Demo of gh-security-config generate](images/gh-security-config-demo.gif)

## Development

To build the extension locally:

```bash
go build -o gh-security-config
```

To run the extension locally:

```bash
./gh-security-config generate
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

This tool is licensed under the MIT License.
