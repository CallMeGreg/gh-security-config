package processors

import (
	"errors"
	"fmt"
	"time"

	"github.com/pterm/pterm"

	"github.com/callmegreg/gh-security-config/internal/types"
)

// SequentialProcessor handles sequential organization processing with optional delay
type SequentialProcessor struct {
	organizations []string
	processor     OrganizationProcessor
	delay         int
	progressBar   *pterm.ProgressbarPrinter
	successCount  int
	skippedCount  int
	errorCount    int
}

// NewSequentialProcessor creates a new sequential processor with optional delay
func NewSequentialProcessor(organizations []string, processor OrganizationProcessor, delay int) *SequentialProcessor {
	return &SequentialProcessor{
		organizations: organizations,
		processor:     processor,
		delay:         delay,
	}
}

// Process executes the organization processing sequentially with optional delay between orgs
func (sp *SequentialProcessor) Process() (successCount, skippedCount, errorCount int) {
	totalOrgs := len(sp.organizations)
	if totalOrgs == 0 {
		return 0, 0, 0
	}

	// Create progress bar
	progressBar, _ := pterm.DefaultProgressbar.WithTotal(totalOrgs).WithTitle("Processing organizations").Start()
	sp.progressBar = progressBar

	// Show delay information if configured
	if sp.delay > 0 {
		pterm.Info.Printf("Processing organizations with %d second delay between each organization\n", sp.delay)
	}

	// Process each organization sequentially
	for i, org := range sp.organizations {
		// Add delay between organizations (not before the first one)
		if i > 0 && sp.delay > 0 {
			// Show loading symbol during delay
			spinner, _ := pterm.DefaultSpinner.WithText(fmt.Sprintf("Waiting %d seconds before processing next organization...", sp.delay)).Start()
			time.Sleep(time.Duration(sp.delay) * time.Second)
			spinner.Success("Ready to process next organization")
		}

		sp.progressBar.UpdateTitle(fmt.Sprintf("Processing %s", org))

		// Process the organization
		result := sp.processor.ProcessOrganization(org)

		if result.Success {
			sp.successCount++
			pterm.Success.Printf("Successfully processed organization '%s'\n", result.Organization)
		} else if result.Skipped {
			sp.skippedCount++
			// Skipped message should already be printed by the processor
		} else if result.Error != nil {
			sp.errorCount++
			// Check if this is a "configuration exists" error
			var configExistsErr *types.ConfigurationExistsError
			if errors.As(result.Error, &configExistsErr) {
				pterm.Warning.Printf("Configuration '%s' already exists in organization '%s', skipping\n", configExistsErr.ConfigName, result.Organization)
				sp.skippedCount++
				sp.errorCount-- // Don't count this as an error
			} else {
				// Check if this is a Dependabot unavailable error (422)
				var dependabotErr *types.DependabotUnavailableError
				if errors.As(result.Error, &dependabotErr) {
					pterm.Error.Printf("Dependabot feature unavailable: %v\n", result.Error)
					pterm.Error.Println("Stopping processing of remaining organizations due to Dependabot unavailability.")
					pterm.Error.Println("Please remove Dependabot settings from your configuration or enable Dependabot on your GHES instance.")

					// Update progress bar to reflect remaining organizations as skipped
					remainingOrgs := totalOrgs - (i + 1)
					sp.skippedCount += remainingOrgs
					sp.progressBar.Add(remainingOrgs)
					sp.progressBar.Stop()

					return sp.successCount, sp.skippedCount, sp.errorCount
				} else {
					pterm.Error.Printf("Failed to process organization '%s': %v\n", result.Organization, result.Error)
				}
			}
		}

		sp.progressBar.Increment()
	}

	progressBar.Stop()
	return sp.successCount, sp.skippedCount, sp.errorCount
}
