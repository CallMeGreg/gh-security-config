package ui

import (
	"fmt"

	"github.com/pterm/pterm"
)

// DisplayCurrentSettings shows current configuration settings with colored output
func DisplayCurrentSettings(settings map[string]interface{}, description string) {
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

// ShowNoOrganizationsWarning displays appropriate warning based on source
func ShowNoOrganizationsWarning(orgListPath string) {
	if orgListPath != "" {
		pterm.Warning.Println("No valid organizations found in the CSV file.")
	} else {
		pterm.Warning.Println("No organizations found in the enterprise.")
	}
}

// ShowOperationCancelled displays cancellation message
func ShowOperationCancelled() {
	pterm.Info.Println("Operation cancelled.")
}

// ShowProcessingStart displays the start of processing with concurrency info
func ShowProcessingStart(orgCount, concurrency int) {
	pterm.Info.Printf("Processing %d organizations with concurrency %d...\n", orgCount, concurrency)
}

// ShowProcessingStartWithDelay displays the start of processing with delay info
func ShowProcessingStartWithDelay(orgCount, delay int) {
	pterm.Info.Printf("Processing %d organizations sequentially with %d second delay between organizations...\n", orgCount, delay)
}
