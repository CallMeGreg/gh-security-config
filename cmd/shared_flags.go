package cmd

import (
	"github.com/spf13/cobra"

	"github.com/callmegreg/gh-security-config/internal/ui"
	"github.com/callmegreg/gh-security-config/internal/utils"
)

// securitySettingFlagNames lists the flag names for each security setting. They are shared
// between the `generate` and `modify` commands.
var securitySettingFlagNames = struct {
	AdvancedSecurity                  string
	DependabotAlerts                  string
	DependabotSecurityUpdates         string
	SecretScanning                    string
	SecretScanningPushProtection      string
	SecretScanningNonProviderPatterns string
	Enforcement                       string
}{
	"advanced-security",
	"dependabot-alerts",
	"dependabot-security-updates",
	"secret-scanning",
	"secret-scanning-push-protection",
	"secret-scanning-non-provider-patterns",
	"enforcement",
}

// addSecuritySettingFlags registers the security-setting flags on the given command. It is
// used by `generate` and `modify` to allow fully non-interactive invocations.
func addSecuritySettingFlags(cmd *cobra.Command) {
	cmd.Flags().String(securitySettingFlagNames.AdvancedSecurity, "", "GitHub Advanced Security setting (enabled, disabled)")
	cmd.Flags().String(securitySettingFlagNames.DependabotAlerts, "", "Dependabot Alerts setting (enabled, disabled, not_set)")
	cmd.Flags().String(securitySettingFlagNames.DependabotSecurityUpdates, "", "Dependabot Security Updates setting (enabled, disabled, not_set)")
	cmd.Flags().String(securitySettingFlagNames.SecretScanning, "", "Secret Scanning setting (enabled, disabled, not_set)")
	cmd.Flags().String(securitySettingFlagNames.SecretScanningPushProtection, "", "Secret Scanning Push Protection setting (enabled, disabled, not_set)")
	cmd.Flags().String(securitySettingFlagNames.SecretScanningNonProviderPatterns, "", "Secret Scanning Non-Provider Patterns setting (enabled, disabled, not_set)")
	cmd.Flags().String(securitySettingFlagNames.Enforcement, "", "Enforcement status for the configuration (enforced, unenforced)")
}

// extractForceFlag reads the universal --force flag. An empty value means "not provided"
// (false). Any other value must be "true" or "false".
func extractForceFlag(cmd *cobra.Command) (bool, error) {
	forceFlag, err := cmd.Flags().GetString("force")
	if err != nil {
		return false, err
	}
	forceOverride, err := utils.ParseBoolStringFlag("force", forceFlag)
	if err != nil {
		return false, err
	}
	if forceOverride == nil {
		return false, nil
	}
	return *forceOverride, nil
}

// extractSecuritySettingOverrides reads each security-setting flag from the command and
// validates it against its allowed set of values. Any flag that is unset returns an empty
// string and triggers an interactive prompt downstream.
func extractSecuritySettingOverrides(cmd *cobra.Command) (ui.SecuritySettingOverrides, error) {
	var out ui.SecuritySettingOverrides

	advSec, err := cmd.Flags().GetString(securitySettingFlagNames.AdvancedSecurity)
	if err != nil {
		return out, err
	}
	if err := utils.ValidateEnumValue(securitySettingFlagNames.AdvancedSecurity, advSec, []string{"enabled", "disabled"}); err != nil {
		return out, err
	}
	out.AdvancedSecurity = advSec

	dbAlerts, err := cmd.Flags().GetString(securitySettingFlagNames.DependabotAlerts)
	if err != nil {
		return out, err
	}
	if err := utils.ValidateEnumValue(securitySettingFlagNames.DependabotAlerts, dbAlerts, []string{"enabled", "disabled", "not_set"}); err != nil {
		return out, err
	}
	out.DependabotAlerts = dbAlerts

	dbSec, err := cmd.Flags().GetString(securitySettingFlagNames.DependabotSecurityUpdates)
	if err != nil {
		return out, err
	}
	if err := utils.ValidateEnumValue(securitySettingFlagNames.DependabotSecurityUpdates, dbSec, []string{"enabled", "disabled", "not_set"}); err != nil {
		return out, err
	}
	out.DependabotSecurityUpdates = dbSec

	ss, err := cmd.Flags().GetString(securitySettingFlagNames.SecretScanning)
	if err != nil {
		return out, err
	}
	if err := utils.ValidateEnumValue(securitySettingFlagNames.SecretScanning, ss, []string{"enabled", "disabled", "not_set"}); err != nil {
		return out, err
	}
	out.SecretScanning = ss

	ssp, err := cmd.Flags().GetString(securitySettingFlagNames.SecretScanningPushProtection)
	if err != nil {
		return out, err
	}
	if err := utils.ValidateEnumValue(securitySettingFlagNames.SecretScanningPushProtection, ssp, []string{"enabled", "disabled", "not_set"}); err != nil {
		return out, err
	}
	out.SecretScanningPushProtection = ssp

	ssnpp, err := cmd.Flags().GetString(securitySettingFlagNames.SecretScanningNonProviderPatterns)
	if err != nil {
		return out, err
	}
	if err := utils.ValidateEnumValue(securitySettingFlagNames.SecretScanningNonProviderPatterns, ssnpp, []string{"enabled", "disabled", "not_set"}); err != nil {
		return out, err
	}
	out.SecretScanningNonProviderPatterns = ssnpp

	enf, err := cmd.Flags().GetString(securitySettingFlagNames.Enforcement)
	if err != nil {
		return out, err
	}
	if err := utils.ValidateEnumValue(securitySettingFlagNames.Enforcement, enf, []string{"enforced", "unenforced"}); err != nil {
		return out, err
	}
	out.Enforcement = enf

	return out, nil
}
