package processors

import (
	"errors"
	"fmt"
	"sync"

	"github.com/pterm/pterm"

	"github.com/callmegreg/gh-security-config/internal/types"
)

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
	stopSignal    chan struct{}
	stopped       bool
}

// NewConcurrentProcessor creates a new concurrent processor
func NewConcurrentProcessor(organizations []string, processor OrganizationProcessor, concurrency int) *ConcurrentProcessor {
	return &ConcurrentProcessor{
		organizations: organizations,
		processor:     processor,
		concurrency:   concurrency,
		stopSignal:    make(chan struct{}),
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
	resultChan := make(chan types.ProcessingResult, totalOrgs)

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

	// Collect results and handle special error cases
	resultsProcessed := 0
	for result := range resultChan {
		cp.mu.Lock()
		resultsProcessed++
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
			var configExistsErr *types.ConfigurationExistsError
			if errors.As(result.Error, &configExistsErr) {
				pterm.Warning.Printf("Configuration '%s' already exists in organization '%s', skipping\n", configExistsErr.ConfigName, result.Organization)
				cp.skippedCount++
				cp.errorCount-- // Don't count this as an error
			} else {
				// Check if this is a Dependabot unavailable error (422)
				var dependabotErr *types.DependabotUnavailableError
				if errors.As(result.Error, &dependabotErr) {
					pterm.Error.Printf("Dependabot feature unavailable: %v\n", result.Error)
					pterm.Error.Println("Stopping processing of remaining organizations due to Dependabot unavailability.")
					pterm.Error.Println("Please remove Dependabot settings from your configuration or enable Dependabot on your GHES instance.")

					// Signal all workers to stop
					if !cp.stopped {
						cp.stopped = true
						close(cp.stopSignal)
					}

					// Update progress bar to reflect remaining organizations as skipped
					remainingOrgs := totalOrgs - resultsProcessed
					cp.skippedCount += remainingOrgs
					cp.progressBar.Add(remainingOrgs)

					cp.mu.Unlock()

					// Drain any remaining results to avoid goroutine leaks
					go func() {
						for range resultChan {
							// Just drain the channel
						}
					}()

					break // Exit the result processing loop
				} else {
					pterm.Error.Printf("Failed to process organization '%s': %v\n", result.Organization, result.Error)
				}
			}
		}

		cp.progressBar.Increment()
		cp.mu.Unlock()
	}

	progressBar.Stop()
	return cp.successCount, cp.skippedCount, cp.errorCount
}

// worker processes organizations from the channel
func (cp *ConcurrentProcessor) worker(wg *sync.WaitGroup, orgChan <-chan string, resultChan chan<- types.ProcessingResult) {
	defer wg.Done()

	for {
		select {
		case org, ok := <-orgChan:
			if !ok {
				return // Channel closed, exit worker
			}
			result := cp.processor.ProcessOrganization(org)
			resultChan <- result
		case <-cp.stopSignal:
			return // Stop signal received, exit worker
		}
	}
}
