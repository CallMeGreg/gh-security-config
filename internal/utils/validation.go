package utils

import "fmt"

// ValidateConcurrency validates the concurrency flag value
func ValidateConcurrency(concurrency int) error {
	if concurrency < 1 || concurrency > 20 {
		return fmt.Errorf("concurrency must be between 1 and 20, got %d", concurrency)
	}
	return nil
}

// ValidateDelay validates the delay flag value
func ValidateDelay(delay int) error {
	if delay < 0 || delay > 600 {
		return fmt.Errorf("delay must be between 1 and 600 seconds, got %d", delay)
	}
	return nil
}

// ValidateConcurrencyAndDelay validates that concurrency and delay are mutually exclusive
func ValidateConcurrencyAndDelay(concurrency, delay int) error {
	// If concurrency is not default (1) and delay is specified, that's an error
	if concurrency > 1 && delay > 0 {
		return fmt.Errorf("--concurrency and --delay flags are mutually exclusive")
	}
	return nil
}
