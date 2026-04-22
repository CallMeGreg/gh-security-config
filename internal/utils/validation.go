package utils

import (
	"fmt"
	"strings"
)

// ValidateEnumValue validates that a value is one of the allowed options.
// Returns an error with a helpful message if not. An empty value is considered valid
// (meaning "not provided") and returns nil.
func ValidateEnumValue(flagName, value string, allowed []string) error {
	if value == "" {
		return nil
	}
	for _, a := range allowed {
		if a == value {
			return nil
		}
	}
	return fmt.Errorf("invalid value for --%s: %q (must be one of: %s)", flagName, value, strings.Join(allowed, ", "))
}

// ParseBoolStringFlag converts a string flag value ("true"/"false"/"") to a *bool.
// An empty string returns nil (meaning "not provided"). Any other value returns an error.
func ParseBoolStringFlag(flagName, value string) (*bool, error) {
	if value == "" {
		return nil, nil
	}
	switch value {
	case "true":
		v := true
		return &v, nil
	case "false":
		v := false
		return &v, nil
	default:
		return nil, fmt.Errorf("invalid value for --%s: %q (must be 'true' or 'false')", flagName, value)
	}
}

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
