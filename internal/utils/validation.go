package utils

import "fmt"

// ValidateConcurrency validates the concurrency flag value
func ValidateConcurrency(concurrency int) error {
	if concurrency < 1 || concurrency > 20 {
		return fmt.Errorf("concurrency must be between 1 and 20, got %d", concurrency)
	}
	return nil
}
