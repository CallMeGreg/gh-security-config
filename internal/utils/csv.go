package utils

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
)

// ReadOrganizationsFromCSV reads organization names from a CSV file
func ReadOrganizationsFromCSV(filePath string) ([]string, error) {
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
