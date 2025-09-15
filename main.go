package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/cli/go-gh/v2"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// Organization represents a GitHub organization
type Organization struct {
	Login string `json:"login"`
}

// SecurityConfiguration represents a GitHub security configuration
type SecurityConfiguration struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ConfigurationExistsError represents an error when a security configuration already exists
type ConfigurationExistsError struct {
	ConfigName string
	OrgName    string
}

func (e *ConfigurationExistsError) Error() string {
	return fmt.Sprintf("configuration '%s' already exists in organization '%s'", e.ConfigName, e.OrgName)
}

// validateConcurrency validates the concurrency flag value
func validateConcurrency(concurrency int) error {
	if concurrency < 1 || concurrency > 20 {
		return fmt.Errorf("concurrency must be between 1 and 20, got %d", concurrency)
	}
	return nil
}

// ProcessingResult represents the result of processing a single organization
type ProcessingResult struct {
	Organization string
	Success      bool
	Skipped      bool
	Error        error
}

// OrganizationProcessor defines the interface for processing organizations
type OrganizationProcessor interface {
	ProcessOrganization(org string) ProcessingResult
}

// ConcurrentProcessor handles concurrent organization processing
type ConcurrentProcessor struct {
	organizations []string
	processor     OrganizationProcessor
	concurrency   int
	progressBar   *pterm.ProgressbarPrinter
	mu            sync.Mutex
	successCount  int
	skippedCount  int
	errorCount    int
}

// NewConcurrentProcessor creates a new concurrent processor
func NewConcurrentProcessor(organizations []string, processor OrganizationProcessor, concurrency int) *ConcurrentProcessor {
	return &ConcurrentProcessor{
		organizations: organizations,
		processor:     processor,
		concurrency:   concurrency,
	}
}

// Process executes the organization processing with the specified concurrency
func (cp *ConcurrentProcessor) Process() (successCount, skippedCount, errorCount int) {
	totalOrgs := len(cp.organizations)
	if totalOrgs == 0 {
		return 0, 0, 0
	}

	// Create progress bar
	progressBar, _ := pterm.DefaultProgressbar.WithTotal(totalOrgs).WithTitle("Processing organizations").Start()
	cp.progressBar = progressBar

	// Create channels for work distribution and result collection
	orgChan := make(chan string, totalOrgs)
	resultChan := make(chan ProcessingResult, totalOrgs)

	// Send all organizations to the work channel
	for _, org := range cp.organizations {
		orgChan <- org
	}
	close(orgChan)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < cp.concurrency; i++ {
		wg.Add(1)
		go cp.worker(&wg, orgChan, resultChan)
	}

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for result := range resultChan {
		cp.mu.Lock()
		cp.progressBar.UpdateTitle(fmt.Sprintf("Processed %s", result.Organization))

		if result.Success {
			cp.successCount++
			pterm.Success.Printf("Successfully processed organization '%s'\n", result.Organization)
		} else if result.Skipped {
			cp.skippedCount++
			// Skipped message should already be printed by the processor
		} else if result.Error != nil {
			cp.errorCount++
			// Check if this is a "configuration exists" error
			var configExistsErr *ConfigurationExistsError
			if errors.As(result.Error, &configExistsErr) {
				pterm.Warning.Printf("Configuration '%s' already exists in organization '%s', skipping\n", configExistsErr.ConfigName, result.Organization)
				cp.skippedCount++
				cp.errorCount-- // Don't count this as an error
			} else {
				pterm.Error.Printf("Failed to process organization '%s': %v\n", result.Organization, result.Error)
			}
		}

		cp.progressBar.Increment()
		cp.mu.Unlock()
	}

	progressBar.Stop()
	return cp.successCount, cp.skippedCount, cp.errorCount
}

// worker processes organizations from the channel
func (cp *ConcurrentProcessor) worker(wg *sync.WaitGroup, orgChan <-chan string, resultChan chan<- ProcessingResult) {
	defer wg.Done()

	for org := range orgChan {
		result := cp.processor.ProcessOrganization(org)
		resultChan <- result
	}
}

// GenerateProcessor implements OrganizationProcessor for the generate command
type GenerateProcessor struct {
	configName        string
	configDescription string
	settings          map[string]interface{}
	scope             string
	setAsDefault      bool
	force             bool
}

// ProcessOrganization processes a single organization for the generate command
func (gp *GenerateProcessor) ProcessOrganization(org string) ProcessingResult {
	// Check membership for this specific organization
	status, err := checkSingleOrganizationMembership(org)
	if err != nil {
		pterm.Warning.Printf("Failed to check membership for organization '%s': %v, skipping\n", org, err)
		return ProcessingResult{Organization: org, Skipped: true}
	}
	if !status.IsMember {
		pterm.Warning.Printf("Skipping organization '%s': You are not a member\n", org)
		return ProcessingResult{Organization: org, Skipped: true}
	}
	if !status.IsOwner {
		pterm.Warning.Printf("Skipping organization '%s': You are a member but not an owner\n", org)
		return ProcessingResult{Organization: org, Skipped: true}
	}

	err = processOrganization(org, gp.configName, gp.configDescription, gp.settings, gp.scope, gp.setAsDefault, gp.force)
	if err != nil {
		return ProcessingResult{Organization: org, Error: err}
	}

	return ProcessingResult{Organization: org, Success: true}
}

// DeleteProcessor implements OrganizationProcessor for the delete command
type DeleteProcessor struct {
	configName string
}

// ProcessOrganization processes a single organization for the delete command
func (dp *DeleteProcessor) ProcessOrganization(org string) ProcessingResult {
	// Check membership for this specific organization
	status, err := checkSingleOrganizationMembership(org)
	if err != nil {
		pterm.Warning.Printf("Failed to check membership for organization '%s': %v, skipping\n", org, err)
		return ProcessingResult{Organization: org, Skipped: true}
	}
	if !status.IsMember {
		pterm.Warning.Printf("Skipping organization '%s': You are not a member\n", org)
		return ProcessingResult{Organization: org, Skipped: true}
	}
	if !status.IsOwner {
		pterm.Warning.Printf("Skipping organization '%s': You are a member but not an owner\n", org)
		return ProcessingResult{Organization: org, Skipped: true}
	}

	deleted, err := deleteConfigurationFromOrg(org, dp.configName)
	if err != nil {
		return ProcessingResult{Organization: org, Error: err}
	}
	if !deleted {
		// Configuration was not found, already logged as warning in deleteConfigurationFromOrg
		return ProcessingResult{Organization: org, Skipped: true}
	}

	return ProcessingResult{Organization: org, Success: true}
}

// ModifyProcessor implements OrganizationProcessor for the modify command
type ModifyProcessor struct {
	configName     string
	newDescription string
	newSettings    map[string]interface{}
}

// ProcessOrganization processes a single organization for the modify command
func (mp *ModifyProcessor) ProcessOrganization(org string) ProcessingResult {
	// Check membership for this specific organization
	status, err := checkSingleOrganizationMembership(org)
	if err != nil {
		pterm.Warning.Printf("Failed to check membership for organization '%s': %v, skipping\n", org, err)
		return ProcessingResult{Organization: org, Skipped: true}
	}
	if !status.IsMember {
		pterm.Warning.Printf("Skipping organization '%s': You are not a member\n", org)
		return ProcessingResult{Organization: org, Skipped: true}
	}
	if !status.IsOwner {
		pterm.Warning.Printf("Skipping organization '%s': You are a member but not an owner\n", org)
		return ProcessingResult{Organization: org, Skipped: true}
	}

	updated, err := modifyConfigurationInOrg(org, mp.configName, mp.newDescription, mp.newSettings)
	if err != nil {
		return ProcessingResult{Organization: org, Error: err}
	}
	if !updated {
		// Configuration was not found, already logged as warning in modifyConfigurationInOrg
		return ProcessingResult{Organization: org, Skipped: true}
	}

	return ProcessingResult{Organization: org, Success: true}
}

var rootCmd = &cobra.Command{
	Use:   "security-config",
	Short: "GitHub Security Configuration Management for Enterprises",
	Long:  "A GitHub CLI extension to manage security configurations across all organizations in an enterprise",
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate and apply security configurations across enterprise organizations",
	Long:  "Interactive command to create security configurations and apply them to organizations in an enterprise. Optionally copy an existing configuration from another organization using --copy-from-org.",
	RunE:  runGenerate,
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete security configurations across enterprise organizations",
	Long:  "Interactive command to delete security configurations from all organizations in an enterprise",
	RunE:  runDelete,
}

var modifyCmd = &cobra.Command{
	Use:   "modify",
	Short: "Modify existing security configurations across enterprise organizations",
	Long:  "Interactive command to update existing security configurations across all organizations in an enterprise",
	RunE:  runModify,
}

func init() {
	generateCmd.Flags().Bool("force", false, "Force deletion of existing configurations with the same name before creating new ones")
	generateCmd.Flags().String("org-list", "", "Path to CSV file containing organization names to target (one per line, no header)")
	generateCmd.Flags().String("copy-from-org", "", "Organization name to copy an existing configuration from")
	generateCmd.Flags().Int("concurrency", 1, "Number of concurrent requests (1-20, default: 1)")
	deleteCmd.Flags().String("org-list", "", "Path to CSV file containing organization names to target (one per line, no header)")
	deleteCmd.Flags().Int("concurrency", 1, "Number of concurrent requests (1-20, default: 1)")
	modifyCmd.Flags().String("org-list", "", "Path to CSV file containing organization names to target (one per line, no header)")
	modifyCmd.Flags().Int("concurrency", 1, "Number of concurrent requests (1-20, default: 1)")
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(modifyCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		pterm.Error.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func runGenerate(cmd *cobra.Command, args []string) error {
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgBlue)).WithTextStyle(pterm.NewStyle(pterm.FgWhite)).Println("GitHub Enterprise Security Configuration Generator")
	pterm.Println()

	// Get org-list flag value
	orgListPath, err := cmd.Flags().GetString("org-list")
	if err != nil {
		return err
	}

	// Validate CSV file early if provided
	if orgListPath != "" {
		orgs, err := readOrganizationsFromCSV(orgListPath)
		if err != nil {
			return fmt.Errorf("CSV validation failed: %w", err)
		}
		if len(orgs) == 0 {
			return fmt.Errorf("CSV file contains no valid organizations")
		}
	}

	// Get force flag value
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	// Get copy-from-org flag value
	copyFromOrg, err := cmd.Flags().GetString("copy-from-org")
	if err != nil {
		return err
	}

	// Get concurrency flag value
	concurrency, err := cmd.Flags().GetInt("concurrency")
	if err != nil {
		return err
	}

	// Validate concurrency
	if err := validateConcurrency(concurrency); err != nil {
		return err
	}

	// Get enterprise name
	enterprise, err := getEnterpriseInput()
	if err != nil {
		return err
	}

	// Get GitHub Enterprise Server URL if needed
	serverURL, err := getServerURLInput()
	if err != nil {
		return err
	}

	// Set hostname if using GitHub Enterprise Server
	if serverURL != "" {
		os.Setenv("GH_HOST", serverURL)
		pterm.Info.Printf("Using GitHub Enterprise Server: %s\n", serverURL)
	}

	// Fetch organizations (from CSV or enterprise API)
	orgs, err := getOrganizations(enterprise, orgListPath)
	if err != nil {
		return err
	}

	if len(orgs) == 0 {
		if orgListPath != "" {
			pterm.Warning.Println("No valid organizations found in the CSV file.")
		} else {
			pterm.Warning.Println("No organizations found in the enterprise.")
		}
		return nil
	}

	var configName, configDescription string
	var settings map[string]interface{}
	var scope string
	var setAsDefault bool

	// Check if we should copy from an existing organization
	if copyFromOrg != "" {
		// Filter out the source organization from target organizations to avoid copying to itself
		var filteredOrgs []string
		for _, org := range orgs {
			if org != copyFromOrg {
				filteredOrgs = append(filteredOrgs, org)
			}
		}

		if len(filteredOrgs) == 0 {
			return fmt.Errorf("no target organizations available after excluding source organization '%s'", copyFromOrg)
		}

		if len(filteredOrgs) < len(orgs) {
			pterm.Info.Printf("Excluding source organization '%s' from targets. Will process %d organizations.\n", copyFromOrg, len(filteredOrgs))
			orgs = filteredOrgs
		}

		// Copy configuration logic
		configName, configDescription, settings, scope, setAsDefault, err = handleCopyFromOrg(copyFromOrg)
		if err != nil {
			return err
		}
	} else {
		// Original logic for creating new configuration
		// Get security configuration details
		configName, configDescription, err = getSecurityConfigInput()
		if err != nil {
			return err
		}

		// Get security settings
		settings, err = getSecuritySettings()
		if err != nil {
			return err
		}

		// Get attachment scope
		scope, err = getAttachmentScope()
		if err != nil {
			return err
		}

		// Ask about setting as default
		setAsDefault, err = getDefaultSetting()
		if err != nil {
			return err
		}
	}

	// Confirm before proceeding
	confirmed, err := confirmOperation(orgs, configName, configDescription, settings, scope, setAsDefault)
	if err != nil {
		return err
	}

	if !confirmed {
		pterm.Info.Println("Operation cancelled.")
		return nil
	}

	// Process each organization
	pterm.Info.Printf("Processing %d organizations with concurrency %d...\n", len(orgs), concurrency)

	// Create processor for generate command
	processor := &GenerateProcessor{
		configName:        configName,
		configDescription: configDescription,
		settings:          settings,
		scope:             scope,
		setAsDefault:      setAsDefault,
		force:             force,
	}

	// Use concurrent processor
	concurrentProcessor := NewConcurrentProcessor(orgs, processor, concurrency)
	successCount, skippedCount, errorCount := concurrentProcessor.Process()

	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgGreen)).WithTextStyle(pterm.NewStyle(pterm.FgBlack)).Printf("Security Configuration Generation Complete! (Success: %d, Skipped: %d, Errors: %d)", successCount, skippedCount, errorCount)

	return nil
}

func runDelete(cmd *cobra.Command, args []string) error {
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgRed)).WithTextStyle(pterm.NewStyle(pterm.FgWhite)).Println("GitHub Enterprise Security Configuration Deletion")
	pterm.Println()

	// Get org-list flag value
	orgListPath, err := cmd.Flags().GetString("org-list")
	if err != nil {
		return err
	}

	// Validate CSV file early if provided
	if orgListPath != "" {
		orgs, err := readOrganizationsFromCSV(orgListPath)
		if err != nil {
			return fmt.Errorf("CSV validation failed: %w", err)
		}
		if len(orgs) == 0 {
			return fmt.Errorf("CSV file contains no valid organizations")
		}
	}

	// Get concurrency flag value
	concurrency, err := cmd.Flags().GetInt("concurrency")
	if err != nil {
		return err
	}

	// Validate concurrency
	if err := validateConcurrency(concurrency); err != nil {
		return err
	}

	// Get enterprise name
	enterprise, err := getEnterpriseInput()
	if err != nil {
		return err
	}

	// Get GitHub Enterprise Server URL if needed
	serverURL, err := getServerURLInput()
	if err != nil {
		return err
	}

	// Set hostname if using GitHub Enterprise Server
	if serverURL != "" {
		os.Setenv("GH_HOST", serverURL)
		pterm.Info.Printf("Using GitHub Enterprise Server: %s\n", serverURL)
	}

	// Fetch organizations (from CSV or enterprise API)
	orgs, err := getOrganizations(enterprise, orgListPath)
	if err != nil {
		return err
	}

	if len(orgs) == 0 {
		if orgListPath != "" {
			pterm.Warning.Println("No valid organizations found in the CSV file.")
		} else {
			pterm.Warning.Println("No organizations found in the enterprise.")
		}
		return nil
	}

	// Get security configuration name to delete
	configName, err := getConfigNameForDeletion()
	if err != nil {
		return err
	}

	// Confirm before proceeding
	confirmed, err := confirmDeleteOperation(orgs, configName)
	if err != nil {
		return err
	}

	if !confirmed {
		pterm.Info.Println("Operation cancelled.")
		return nil
	}

	// Process each organization
	pterm.Info.Printf("Processing %d organizations with concurrency %d...\n", len(orgs), concurrency)

	// Create processor for delete command
	processor := &DeleteProcessor{
		configName: configName,
	}

	// Use concurrent processor
	concurrentProcessor := NewConcurrentProcessor(orgs, processor, concurrency)
	successCount, skippedCount, errorCount := concurrentProcessor.Process()

	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgGreen)).WithTextStyle(pterm.NewStyle(pterm.FgBlack)).Printf("Security Configuration Deletion Complete! (Success: %d, Skipped: %d, Errors: %d)", successCount, skippedCount, errorCount)

	return nil
}

func runModify(cmd *cobra.Command, args []string) error {
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgMagenta)).WithTextStyle(pterm.NewStyle(pterm.FgWhite)).Println("GitHub Enterprise Security Configuration Modification")
	pterm.Println()

	// Get org-list flag value
	orgListPath, err := cmd.Flags().GetString("org-list")
	if err != nil {
		return err
	}

	// Validate CSV file early if provided
	if orgListPath != "" {
		orgs, err := readOrganizationsFromCSV(orgListPath)
		if err != nil {
			return fmt.Errorf("CSV validation failed: %w", err)
		}
		if len(orgs) == 0 {
			return fmt.Errorf("CSV file contains no valid organizations")
		}
	}

	// Get concurrency flag value
	concurrency, err := cmd.Flags().GetInt("concurrency")
	if err != nil {
		return err
	}

	// Validate concurrency
	if err := validateConcurrency(concurrency); err != nil {
		return err
	}

	// Get enterprise name
	enterprise, err := getEnterpriseInput()
	if err != nil {
		return err
	}

	// Get GitHub Enterprise Server URL if needed
	serverURL, err := getServerURLInput()
	if err != nil {
		return err
	}

	// Set hostname if using GitHub Enterprise Server
	if serverURL != "" {
		os.Setenv("GH_HOST", serverURL)
		pterm.Info.Printf("Using GitHub Enterprise Server: %s\n", serverURL)
	}

	// Fetch organizations (from CSV or enterprise API)
	orgs, err := getOrganizations(enterprise, orgListPath)
	if err != nil {
		return err
	}

	if len(orgs) == 0 {
		if orgListPath != "" {
			pterm.Warning.Println("No valid organizations found in the CSV file.")
		} else {
			pterm.Warning.Println("No organizations found in the enterprise.")
		}
		return nil
	}

	// Get security configuration name to modify
	configName, err := getConfigNameForModification()
	if err != nil {
		return err
	}

	// Fetch existing configuration details from first accessible organization to show current settings
	var currentSettings map[string]interface{}
	var currentDescription string
	for _, org := range orgs {
		// Check membership for this specific organization
		status, err := checkSingleOrganizationMembership(org)
		if err != nil || !status.IsMember || !status.IsOwner {
			continue
		}

		configs, err := fetchSecurityConfigurations(org)
		if err != nil {
			continue
		}

		configID, found := findConfigurationByName(configs, configName)
		if found {
			// Get detailed configuration
			configDetails, err := getSecurityConfigurationDetails(org, configID)
			if err == nil {
				currentSettings = configDetails.Settings
				currentDescription = configDetails.Description
				break
			}
		}
	}

	if currentSettings == nil {
		pterm.Warning.Printf("Configuration '%s' not found in any accessible organizations.\n", configName)
		return fmt.Errorf("configuration '%s' not found", configName)
	}

	// Show current settings and get new settings
	pterm.Info.Println("Current configuration settings:")
	displayCurrentSettings(currentSettings, currentDescription)
	pterm.Println()

	// Get new description
	newDescription, err := getUpdatedDescription(currentDescription)
	if err != nil {
		return err
	}

	// Get updated security settings
	newSettings, err := getSecuritySettingsForUpdate(currentSettings)
	if err != nil {
		return err
	}

	// Confirm before proceeding
	confirmed, err := confirmModifyOperation(orgs, configName, currentDescription, newDescription, currentSettings, newSettings)
	if err != nil {
		return err
	}

	if !confirmed {
		pterm.Info.Println("Operation cancelled.")
		return nil
	}

	// Process each organization
	pterm.Info.Printf("Processing %d organizations with concurrency %d...\n", len(orgs), concurrency)

	// Create processor for modify command
	processor := &ModifyProcessor{
		configName:     configName,
		newDescription: newDescription,
		newSettings:    newSettings,
	}

	// Use concurrent processor
	concurrentProcessor := NewConcurrentProcessor(orgs, processor, concurrency)
	successCount, skippedCount, errorCount := concurrentProcessor.Process()

	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgGreen)).WithTextStyle(pterm.NewStyle(pterm.FgBlack)).Printf("Security Configuration Modification Complete! (Success: %d, Skipped: %d, Errors: %d)", successCount, skippedCount, errorCount)

	return nil
}

func getEnterpriseInput() (string, error) {
	enterprise, err := pterm.DefaultInteractiveTextInput.WithDefaultText("").WithMultiLine(false).Show("Enter the enterprise slug (e.g., github)")
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(enterprise) == "" {
		return "", fmt.Errorf("enterprise slug is required")
	}

	return strings.TrimSpace(enterprise), nil
}

func getServerURLInput() (string, error) {
	isGHES, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Are you using GitHub Enterprise Server (not GitHub.com)?").WithDefaultValue(true).Show()
	if err != nil {
		return "", err
	}

	if !isGHES {
		return "", nil
	}

	serverURL, err := pterm.DefaultInteractiveTextInput.WithDefaultText("").WithMultiLine(false).Show("Enter your GitHub Enterprise Server URL (e.g., github.company.com)")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(serverURL), nil
}

func getSecurityConfigInput() (string, string, error) {
	name, err := pterm.DefaultInteractiveTextInput.WithDefaultText("Enterprise Security Configuration").WithMultiLine(false).Show("Enter security configuration name")
	if err != nil {
		return "", "", err
	}

	description, err := pterm.DefaultInteractiveTextInput.WithDefaultText("Security configuration applied across enterprise organizations").WithMultiLine(false).Show("Enter security configuration description")
	if err != nil {
		return "", "", err
	}

	return strings.TrimSpace(name), strings.TrimSpace(description), nil
}

func getSecuritySettings() (map[string]interface{}, error) {
	settings := make(map[string]interface{})

	pterm.Info.Println("Configure security settings:")

	// Advanced Security
	advancedSecurity, err := pterm.DefaultInteractiveSelect.WithOptions([]string{"enabled", "disabled"}).WithDefaultOption("enabled").Show("GitHub Advanced Security")
	if err != nil {
		return nil, err
	}
	settings["advanced_security"] = advancedSecurity

	// Secret Scanning
	secretScanning, err := pterm.DefaultInteractiveSelect.WithOptions([]string{"enabled", "disabled", "not_set"}).WithDefaultOption("enabled").Show("Secret Scanning")
	if err != nil {
		return nil, err
	}
	settings["secret_scanning"] = secretScanning

	// Secret Scanning Push Protection
	pushProtection, err := pterm.DefaultInteractiveSelect.WithOptions([]string{"enabled", "disabled", "not_set"}).WithDefaultOption("enabled").Show("Secret Scanning Push Protection")
	if err != nil {
		return nil, err
	}
	settings["secret_scanning_push_protection"] = pushProtection

	// Secret Scanning Non-Provider Patterns
	nonProviderPatterns, err := pterm.DefaultInteractiveSelect.WithOptions([]string{"enabled", "disabled", "not_set"}).WithDefaultOption("disabled").Show("Secret Scanning Non-Provider Patterns")
	if err != nil {
		return nil, err
	}
	settings["secret_scanning_non_provider_patterns"] = nonProviderPatterns

	// Enforcement
	enforcement, err := pterm.DefaultInteractiveSelect.WithOptions([]string{"enforced", "unenforced"}).WithDefaultOption("enforced").Show("Enforcement Status")
	if err != nil {
		return nil, err
	}
	settings["enforcement"] = enforcement

	return settings, nil
}

func getAttachmentScope() (string, error) {
	scope, err := pterm.DefaultInteractiveSelect.WithOptions([]string{
		"all",
		"public",
		"private_or_internal",
	}).WithDefaultOption("all").Show("Select repositories to attach configuration to")
	if err != nil {
		return "", err
	}

	return scope, nil
}

func getDefaultSetting() (bool, error) {
	setDefault, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Set this configuration as default for new repositories?").Show()
	if err != nil {
		return false, err
	}

	return setDefault, nil
}

func confirmOperation(orgs []string, configName, configDescription string, settings map[string]interface{}, scope string, setAsDefault bool) (bool, error) {
	pterm.Println()
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgYellow)).WithTextStyle(pterm.NewStyle(pterm.FgBlack)).Println("Operation Summary")

	pterm.Printf("Organizations: %d\n", len(orgs))
	pterm.Printf("Configuration Name: %s\n", pterm.Yellow(configName))
	pterm.Printf("Configuration Description: %s\n", pterm.Yellow(configDescription))
	pterm.Println()

	pterm.Info.Println("Security Settings:")
	for key, value := range settings {
		valueStr := fmt.Sprintf("%v", value)
		var coloredValue string

		switch valueStr {
		case "enabled", "enforced":
			coloredValue = pterm.Green(valueStr)
		case "disabled", "unenforced":
			coloredValue = pterm.Red(valueStr)
		case "not_set":
			coloredValue = pterm.Yellow(valueStr)
		default:
			coloredValue = pterm.Yellow(valueStr)
		}

		pterm.Printf("  %s: %s\n", pterm.Cyan(key), coloredValue)
	}
	pterm.Println()

	pterm.Printf("Attachment Scope: %s\n", pterm.Magenta(scope))
	pterm.Printf("Set as Default: %s\n", pterm.Cyan(fmt.Sprintf("%t", setAsDefault)))
	pterm.Println()

	confirmed, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Proceed with creating security configurations?").Show()
	if err != nil {
		return false, err
	}

	return confirmed, nil
}

func fetchOrganizations(enterprise string) ([]string, error) {
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

func formatCursor(cursor *string) string {
	if cursor == nil {
		return "null"
	}
	return fmt.Sprintf(`"%s"`, *cursor)
}

// readOrganizationsFromCSV reads organization names from a CSV file
func readOrganizationsFromCSV(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %w", err)
	}

	var orgs []string
	for i, record := range records {
		if len(record) == 0 {
			continue // Skip empty lines
		}
		orgName := strings.TrimSpace(record[0])
		if orgName == "" {
			continue // Skip empty organization names
		}
		// Basic validation for organization name format
		if strings.Contains(orgName, " ") || strings.Contains(orgName, "/") {
			pterm.Warning.Printf("Line %d: Invalid organization name format '%s', skipping\n", i+1, orgName)
			continue
		}
		orgs = append(orgs, orgName)
	}

	return orgs, nil
}

// getOrganizations returns organization list either from CSV file or from enterprise API
func getOrganizations(enterprise, orgListPath string) ([]string, error) {
	if orgListPath != "" {
		pterm.Info.Printf("Reading organizations from CSV file: %s\n", orgListPath)
		csvOrgs, err := readOrganizationsFromCSV(orgListPath)
		if err != nil {
			return nil, err
		}
		if len(csvOrgs) == 0 {
			return nil, fmt.Errorf("no valid organizations found in CSV file")
		}
		pterm.Success.Printf("Found %d organizations in CSV file\n", len(csvOrgs))

		// Fetch all organizations from enterprise to validate against
		pterm.Info.Println("Fetching organizations from enterprise to validate CSV list...")
		enterpriseOrgs, err := fetchOrganizations(enterprise)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch enterprise organizations for validation: %w", err)
		}
		pterm.Success.Printf("Found %d organizations in enterprise '%s'\n", len(enterpriseOrgs), enterprise)

		// Create a map for faster lookup
		enterpriseOrgMap := make(map[string]bool)
		for _, org := range enterpriseOrgs {
			enterpriseOrgMap[org] = true
		}

		// Validate CSV organizations against enterprise organizations
		var validOrgs []string
		var invalidOrgs []string

		for _, org := range csvOrgs {
			if enterpriseOrgMap[org] {
				validOrgs = append(validOrgs, org)
			} else {
				invalidOrgs = append(invalidOrgs, org)
			}
		}

		// Warn about invalid organizations
		if len(invalidOrgs) > 0 {
			pterm.Warning.Printf("Found %d organizations in CSV that do not exist in enterprise '%s':\n", len(invalidOrgs), enterprise)
			for _, org := range invalidOrgs {
				pterm.Printf("  - %s (not found in enterprise)\n", pterm.Red(org))
			}
			pterm.Println()
		}

		// Check if we have any valid organizations left
		if len(validOrgs) == 0 {
			return nil, fmt.Errorf("no valid organizations found in CSV file that exist in enterprise '%s'", enterprise)
		}

		if len(invalidOrgs) > 0 {
			pterm.Info.Printf("Proceeding with %d valid organizations (skipping %d invalid)\n", len(validOrgs), len(invalidOrgs))
		}

		// Show the list of valid organizations that will be targeted
		pterm.Info.Println("Valid organizations to be targeted:")
		for _, org := range validOrgs {
			pterm.Printf("  - %s\n", pterm.Green(org))
		}
		pterm.Println()

		return validOrgs, nil
	}

	// Use existing enterprise API fetching
	pterm.Info.Println("Fetching organizations from enterprise...")
	orgs, err := fetchOrganizations(enterprise)
	if err != nil {
		return nil, err
	}
	pterm.Success.Printf("Found %d organizations in enterprise '%s'\n", len(orgs), enterprise)
	return orgs, nil
}

// MembershipStatus represents the user's membership status in an organization
type MembershipStatus struct {
	IsMember bool
	IsOwner  bool
	Role     string
}

func getCurrentUser() (string, error) {
	userResponse, _, err := gh.Exec("api", "user", "-q", ".login")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(userResponse.String()), nil
}

func checkSingleOrganizationMembership(org string) (MembershipStatus, error) {
	// Get current user's login first
	currentUser, err := getCurrentUser()
	if err != nil {
		return MembershipStatus{}, fmt.Errorf("failed to get current user: %w", err)
	}

	// Use REST API to check membership and role directly
	userResponse, stderr, err := gh.Exec("api", "-H", "Accept: application/vnd.github+json", "-H", "X-GitHub-Api-Version: 2022-11-28", fmt.Sprintf("/orgs/%s/memberships/%s", org, currentUser))
	if err != nil {
		// If we get a 404 or similar error, the user is likely not a member
		if strings.Contains(stderr.String(), "404") || strings.Contains(stderr.String(), "Not Found") {
			return MembershipStatus{IsMember: false, IsOwner: false, Role: "none"}, nil
		}
		return MembershipStatus{IsMember: false, IsOwner: false, Role: "none"}, nil
	}

	var membership struct {
		State string `json:"state"`
		Role  string `json:"role"`
	}

	if err := json.Unmarshal(userResponse.Bytes(), &membership); err != nil {
		pterm.Warning.Printf("Failed to parse membership data for organization '%s': %v\n", org, err)
		return MembershipStatus{IsMember: false, IsOwner: false, Role: "none"}, nil
	}

	// Check if membership is active and determine role
	if membership.State == "active" {
		isOwner := membership.Role == "admin"
		return MembershipStatus{
			IsMember: true,
			IsOwner:  isOwner,
			Role:     membership.Role,
		}, nil
	}

	// If state is not active, treat as not a member
	return MembershipStatus{IsMember: false, IsOwner: false, Role: "none"}, nil
}

func processOrganization(org, configName, configDescription string, settings map[string]interface{}, scope string, setAsDefault bool, force bool) error {
	// Check if a configuration with the same name already exists
	configs, err := fetchSecurityConfigurations(org)
	if err != nil {
		return fmt.Errorf("failed to fetch existing security configurations: %w", err)
	}

	// Check if configuration already exists
	existingConfigID, exists := findConfigurationByName(configs, configName)
	if exists {
		if force {
			// Delete the existing configuration
			pterm.Info.Printf("Force flag enabled: deleting existing configuration '%s' from organization '%s'\n", configName, org)
			err = deleteSecurityConfiguration(org, existingConfigID)
			if err != nil {
				return fmt.Errorf("failed to delete existing security configuration: %w", err)
			}
		} else {
			return &ConfigurationExistsError{
				ConfigName: configName,
				OrgName:    org,
			}
		}
	}

	// Create security configuration
	configID, err := createSecurityConfiguration(org, configName, configDescription, settings)
	if err != nil {
		return fmt.Errorf("failed to create security configuration: %w", err)
	}

	// Attach configuration to repositories
	err = attachConfigurationToRepos(org, configID, scope)
	if err != nil {
		return fmt.Errorf("failed to attach configuration to repositories: %w", err)
	}

	// Set as default if requested
	if setAsDefault {
		err = setConfigurationAsDefault(org, configID)
		if err != nil {
			return fmt.Errorf("failed to set configuration as default: %w", err)
		}
	}

	return nil
}

func createSecurityConfiguration(org, name, description string, settings map[string]interface{}) (int, error) {
	// Build the request body
	body := map[string]interface{}{
		"name":        name,
		"description": description,
	}

	// Add all settings to the body
	for key, value := range settings {
		body[key] = value
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	// Create temporary file for the JSON body
	tmpFile, err := os.CreateTemp("", "security-config-*.json")
	if err != nil {
		return 0, err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(bodyBytes); err != nil {
		return 0, err
	}
	tmpFile.Close()

	// Execute the gh API command
	response, stderr, err := gh.Exec("api", "--method", "POST", "-H", "Accept: application/vnd.github+json", "-H", "X-GitHub-Api-Version: 2022-11-28", fmt.Sprintf("/orgs/%s/code-security/configurations", org), "--input", tmpFile.Name())
	if err != nil {
		pterm.Error.Printf("Failed to create security configuration for org '%s': %v\n", org, err)
		pterm.Error.Printf("gh CLI stderr: %s\n", stderr.String())
		return 0, err
	}

	var config SecurityConfiguration
	if err := json.Unmarshal(response.Bytes(), &config); err != nil {
		return 0, err
	}

	return config.ID, nil
}

func attachConfigurationToRepos(org string, configID int, scope string) error {
	body := map[string]interface{}{
		"scope": scope,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// Create temporary file for the JSON body
	tmpFile, err := os.CreateTemp("", "attach-config-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(bodyBytes); err != nil {
		return err
	}
	tmpFile.Close()

	_, _, err = gh.Exec("api", "--method", "POST", "-H", "Accept: application/vnd.github+json", "-H", "X-GitHub-Api-Version: 2022-11-28", fmt.Sprintf("/orgs/%s/code-security/configurations/%d/attach", org, configID), "--input", tmpFile.Name())
	return err
}

func setConfigurationAsDefault(org string, configID int) error {
	body := map[string]interface{}{
		"default_for_new_repos": "all",
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// Create temporary file for the JSON body
	tmpFile, err := os.CreateTemp("", "default-config-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(bodyBytes); err != nil {
		return err
	}
	tmpFile.Close()

	_, _, err = gh.Exec("api", "--method", "PUT", "-H", "Accept: application/vnd.github+json", "-H", "X-GitHub-Api-Version: 2022-11-28", fmt.Sprintf("/orgs/%s/code-security/configurations/%d/defaults", org, configID), "--input", tmpFile.Name())
	return err
}

// Helper functions for modify functionality

type SecurityConfigurationDetails struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Settings    map[string]interface{} `json:"-"` // Will be populated separately
}

func getConfigNameForModification() (string, error) {
	configName, err := pterm.DefaultInteractiveTextInput.WithDefaultText("").WithMultiLine(false).Show("Enter the name of the security configuration to modify")
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(configName) == "" {
		return "", fmt.Errorf("configuration name is required")
	}

	return strings.TrimSpace(configName), nil
}

func getSecurityConfigurationDetails(org string, configID int) (*SecurityConfigurationDetails, error) {
	response, stderr, err := gh.Exec("api", "-H", "Accept: application/vnd.github+json", "-H", "X-GitHub-Api-Version: 2022-11-28", fmt.Sprintf("/orgs/%s/code-security/configurations/%d", org, configID))
	if err != nil {
		pterm.Error.Printf("Failed to fetch security configuration details for org '%s': %v\n", org, err)
		pterm.Error.Printf("gh CLI stderr: %s\n", stderr.String())
		return nil, err
	}

	var configResponse map[string]interface{}
	if err := json.Unmarshal(response.Bytes(), &configResponse); err != nil {
		return nil, err
	}

	details := &SecurityConfigurationDetails{
		Settings: make(map[string]interface{}),
	}

	// Extract basic info
	if id, ok := configResponse["id"].(float64); ok {
		details.ID = int(id)
	}
	if name, ok := configResponse["name"].(string); ok {
		details.Name = name
	}
	if desc, ok := configResponse["description"].(string); ok {
		details.Description = desc
	}

	// Extract security settings
	securitySettings := []string{
		"advanced_security", "secret_scanning", "secret_scanning_push_protection",
		"secret_scanning_non_provider_patterns", "enforcement",
	}

	for _, setting := range securitySettings {
		if val, exists := configResponse[setting]; exists {
			details.Settings[setting] = val
		}
	}

	return details, nil
}

func displayCurrentSettings(settings map[string]interface{}, description string) {
	pterm.Printf("  Description: %s\n", pterm.Yellow(description))
	for key, value := range settings {
		valueStr := fmt.Sprintf("%v", value)
		var coloredValue string

		switch valueStr {
		case "enabled", "enforced":
			coloredValue = pterm.Green(valueStr)
		case "disabled", "unenforced":
			coloredValue = pterm.Red(valueStr)
		case "not_set":
			coloredValue = pterm.Yellow(valueStr)
		default:
			coloredValue = pterm.Yellow(valueStr)
		}

		pterm.Printf("  %s: %s\n", pterm.Cyan(key), coloredValue)
	}
}

func getUpdatedDescription(currentDescription string) (string, error) {
	newDescription, err := pterm.DefaultInteractiveTextInput.WithDefaultText(currentDescription).WithMultiLine(false).Show("Enter updated security configuration description")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(newDescription), nil
}

func getSecuritySettingsForUpdate(currentSettings map[string]interface{}) (map[string]interface{}, error) {
	newSettings := make(map[string]interface{})

	pterm.Info.Println("Update security settings (press Enter to keep current value):")

	settingsConfig := []struct {
		key          string
		description  string
		options      []string
		defaultValue string
	}{
		{"advanced_security", "GitHub Advanced Security", []string{"enabled", "disabled"}, "enabled"},
		{"secret_scanning", "Secret Scanning", []string{"enabled", "disabled", "not_set"}, "enabled"},
		{"secret_scanning_push_protection", "Secret Scanning Push Protection", []string{"enabled", "disabled", "not_set"}, "enabled"},
		{"secret_scanning_non_provider_patterns", "Secret Scanning Non-Provider Patterns", []string{"enabled", "disabled", "not_set"}, "disabled"},
		{"enforcement", "Enforcement Status", []string{"enforced", "unenforced"}, "enforced"},
	}

	for _, config := range settingsConfig {
		currentValue := "not_set"
		if val, exists := currentSettings[config.key]; exists {
			currentValue = fmt.Sprintf("%v", val)
		}

		// Add option to keep current value
		options := append([]string{fmt.Sprintf("Keep current (%s)", currentValue)}, config.options...)

		selection, err := pterm.DefaultInteractiveSelect.WithOptions(options).WithDefaultOption(options[0]).Show(config.description)
		if err != nil {
			return nil, err
		}

		// If user chose to keep current value, use the current value
		if strings.HasPrefix(selection, "Keep current") {
			newSettings[config.key] = currentValue
		} else {
			newSettings[config.key] = selection
		}
	}

	return newSettings, nil
}

func confirmModifyOperation(orgs []string, configName, currentDescription, newDescription string, currentSettings, newSettings map[string]interface{}) (bool, error) {
	pterm.Println()
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgYellow)).WithTextStyle(pterm.NewStyle(pterm.FgBlack)).Println("MODIFY OPERATION SUMMARY")

	pterm.Printf("Organizations: %d\n", len(orgs))
	pterm.Printf("Configuration to Modify: %s\n", pterm.Magenta(configName))
	pterm.Println()

	// Show changes
	pterm.Info.Println("Changes to be made:")

	// Description changes
	if currentDescription != newDescription {
		pterm.Printf("  Description: %s → %s\n", pterm.Red(currentDescription), pterm.Green(newDescription))
	} else {
		pterm.Printf("  Description: %s (no change)\n", pterm.Yellow(currentDescription))
	}

	// Setting changes
	for key, newValue := range newSettings {
		currentValue := fmt.Sprintf("%v", currentSettings[key])
		newValueStr := fmt.Sprintf("%v", newValue)

		if currentValue != newValueStr {
			pterm.Printf("  %s: %s → %s\n", pterm.Cyan(key), pterm.Red(currentValue), pterm.Green(newValueStr))
		} else {
			pterm.Printf("  %s: %s (no change)\n", pterm.Cyan(key), pterm.Yellow(currentValue))
		}
	}

	pterm.Println()

	confirmed, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Proceed with modifying security configurations?").Show()
	if err != nil {
		return false, err
	}

	return confirmed, nil
}

func modifyConfigurationInOrg(org, configName, newDescription string, newSettings map[string]interface{}) (bool, error) {
	// First, fetch security configurations for the organization
	configs, err := fetchSecurityConfigurations(org)
	if err != nil {
		return false, fmt.Errorf("failed to fetch security configurations: %w", err)
	}

	// Find the configuration by name
	configID, found := findConfigurationByName(configs, configName)
	if !found {
		pterm.Warning.Printf("Configuration '%s' not found in organization '%s', skipping\n", configName, org)
		return false, nil // Not an error, just skip this org
	}

	// Update the configuration
	err = updateSecurityConfiguration(org, configID, newDescription, newSettings)
	if err != nil {
		return false, fmt.Errorf("failed to update security configuration: %w", err)
	}

	return true, nil
}

func updateSecurityConfiguration(org string, configID int, description string, settings map[string]interface{}) error {
	// Build the request body for PATCH request
	body := map[string]interface{}{
		"description": description,
	}

	// Add all settings to the body
	for key, value := range settings {
		body[key] = value
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// Create temporary file for the JSON body
	tmpFile, err := os.CreateTemp("", "update-config-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(bodyBytes); err != nil {
		return err
	}
	tmpFile.Close()

	// Execute the gh API command with PATCH method
	_, stderr, err := gh.Exec("api", "--method", "PATCH", "-H", "Accept: application/vnd.github+json", "-H", "X-GitHub-Api-Version: 2022-11-28", fmt.Sprintf("/orgs/%s/code-security/configurations/%d", org, configID), "--input", tmpFile.Name())
	if err != nil {
		pterm.Error.Printf("Failed to update security configuration %d for org '%s': %v\n", configID, org, err)
		pterm.Error.Printf("gh CLI stderr: %s\n", stderr.String())
		return err
	}

	return nil
}

// Helper functions for delete functionality

func getConfigNameForDeletion() (string, error) {
	configName, err := pterm.DefaultInteractiveTextInput.WithDefaultText("").WithMultiLine(false).Show("Enter the name of the security configuration to delete")
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(configName) == "" {
		return "", fmt.Errorf("configuration name is required")
	}

	return strings.TrimSpace(configName), nil
}

func confirmDeleteOperation(orgs []string, configName string) (bool, error) {
	pterm.Println()
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgRed)).WithTextStyle(pterm.NewStyle(pterm.FgWhite)).Println("DELETE OPERATION SUMMARY")

	pterm.Printf("Organizations: %d\n", len(orgs))
	pterm.Printf("Configuration to Delete: %s\n", pterm.Red(configName))
	pterm.Println()

	pterm.Warning.Println("WARNING: This operation will delete the security configuration from ALL organizations in the enterprise.")
	pterm.Warning.Println("This action cannot be undone. Repositories will retain their settings but will no longer be associated with the configuration.")
	pterm.Println()

	confirmed, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Are you absolutely sure you want to proceed with deleting this configuration?").WithDefaultValue(false).Show()
	if err != nil {
		return false, err
	}

	return confirmed, nil
}

func deleteConfigurationFromOrg(org, configName string) (bool, error) {
	// First, fetch security configurations for the organization
	configs, err := fetchSecurityConfigurations(org)
	if err != nil {
		return false, fmt.Errorf("failed to fetch security configurations: %w", err)
	}

	// Find the configuration by name
	configID, found := findConfigurationByName(configs, configName)
	if !found {
		pterm.Warning.Printf("Configuration '%s' not found in organization '%s', skipping\n", configName, org)
		return false, nil // Not an error, just skip this org
	}

	// Delete the configuration
	err = deleteSecurityConfiguration(org, configID)
	if err != nil {
		return false, fmt.Errorf("failed to delete security configuration: %w", err)
	}

	return true, nil
}

func fetchSecurityConfigurations(org string) ([]SecurityConfiguration, error) {
	response, stderr, err := gh.Exec("api", "-H", "Accept: application/vnd.github+json", "-H", "X-GitHub-Api-Version: 2022-11-28", fmt.Sprintf("/orgs/%s/code-security/configurations", org))
	if err != nil {
		pterm.Error.Printf("Failed to fetch security configurations for org '%s': %v\n", org, err)
		pterm.Error.Printf("gh CLI stderr: %s\n", stderr.String())
		return nil, err
	}

	var configs []SecurityConfiguration
	if err := json.Unmarshal(response.Bytes(), &configs); err != nil {
		return nil, err
	}

	return configs, nil
}

// handleCopyFromOrg handles the logic for copying a configuration from an existing organization
func handleCopyFromOrg(copyFromOrg string) (string, string, map[string]interface{}, string, bool, error) {
	pterm.Info.Printf("Fetching security configurations from organization '%s'...\n", copyFromOrg)

	// Check if user has access to the source organization
	status, err := checkSingleOrganizationMembership(copyFromOrg)
	if err != nil {
		return "", "", nil, "", false, fmt.Errorf("failed to check membership for organization '%s': %w", copyFromOrg, err)
	}
	if !status.IsMember {
		return "", "", nil, "", false, fmt.Errorf("you are not a member of organization '%s'", copyFromOrg)
	}
	if !status.IsOwner {
		return "", "", nil, "", false, fmt.Errorf("you are a member but not an owner of organization '%s'", copyFromOrg)
	}

	// Fetch security configurations from the source organization
	configs, err := fetchSecurityConfigurations(copyFromOrg)
	if err != nil {
		return "", "", nil, "", false, fmt.Errorf("failed to fetch security configurations from organization '%s': %w", copyFromOrg, err)
	}

	if len(configs) == 0 {
		return "", "", nil, "", false, fmt.Errorf("no security configurations found in organization '%s'", copyFromOrg)
	}

	// Present configurations for selection
	var configOptions []string
	configMap := make(map[string]SecurityConfiguration)
	for _, config := range configs {
		displayName := fmt.Sprintf("%s - %s", config.Name, config.Description)
		configOptions = append(configOptions, displayName)
		configMap[displayName] = config
	}

	selectedConfig, err := pterm.DefaultInteractiveSelect.WithOptions(configOptions).Show("Select a configuration to copy")
	if err != nil {
		return "", "", nil, "", false, err
	}

	// Get the selected configuration details
	selectedConfigData := configMap[selectedConfig]

	// Get detailed configuration including settings
	configDetails, err := getSecurityConfigurationDetails(copyFromOrg, selectedConfigData.ID)
	if err != nil {
		return "", "", nil, "", false, fmt.Errorf("failed to fetch configuration details: %w", err)
	}

	pterm.Success.Printf("Selected configuration '%s' from organization '%s'\n", selectedConfigData.Name, copyFromOrg)

	// Display current settings
	pterm.Info.Println("Configuration details that will be copied:")
	displayCurrentSettings(configDetails.Settings, configDetails.Description)
	pterm.Println()

	// Ask for attachment scope (this might be different for target organizations)
	scope, err := getAttachmentScope()
	if err != nil {
		return "", "", nil, "", false, err
	}

	// Ask about setting as default (this might be different for target organizations)
	setAsDefault, err := getDefaultSetting()
	if err != nil {
		return "", "", nil, "", false, err
	}

	return selectedConfigData.Name, configDetails.Description, configDetails.Settings, scope, setAsDefault, nil
}

func findConfigurationByName(configs []SecurityConfiguration, name string) (int, bool) {
	for _, config := range configs {
		if config.Name == name {
			return config.ID, true
		}
	}
	return 0, false
}

func deleteSecurityConfiguration(org string, configID int) error {
	_, stderr, err := gh.Exec("api", "--method", "DELETE", "-H", "Accept: application/vnd.github+json", "-H", "X-GitHub-Api-Version: 2022-11-28", fmt.Sprintf("/orgs/%s/code-security/configurations/%d", org, configID))
	if err != nil {
		pterm.Error.Printf("Failed to delete security configuration %d from org '%s': %v\n", configID, org, err)
		pterm.Error.Printf("gh CLI stderr: %s\n", stderr.String())
		return err
	}

	return nil
}
