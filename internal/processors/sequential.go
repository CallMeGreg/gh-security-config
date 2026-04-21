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

	// Process each organization sequentially
	for i, org := range sp.organizations {
		// Add delay between organizations (not before the first one)
		if i > 0 && sp.delay > 0 {
			for remaining := sp.delay; remaining > 0; remaining-- {
				sp.progressBar.UpdateTitle(fmt.Sprintf("Waiting %d seconds before processing next organization...", remaining))
				time.Sleep(time.Second)
			}
		}

		// Increment before processing so the bar shows 1-based progress (e.g. "1/5")
		sp.progressBar.Increment()
		sp.progressBar.UpdateTitle(fmt.Sprintf("Processing %s", org))

		// Process the organization
		result := sp.processor.ProcessOrganization(org)

		if result.Success {
			sp.successCount++
			sp.progressBar.UpdateTitle(fmt.Sprintf("Processed %s", result.Organization))
		} else if result.Skipped {
			sp.skippedCount++
			sp.progressBar.UpdateTitle(fmt.Sprintf("Skipped %s", result.Organization))
		} else if result.Error != nil {
			sp.errorCount++
			sp.progressBar.UpdateTitle(fmt.Sprintf("Processed %s", result.Organization))
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

					// Add remaining orgs as skipped
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

	}

	progressBar.Stop()
	return sp.successCount, sp.skippedCount, sp.errorCount
}
