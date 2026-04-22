package utils

import (
	"strings"
	"testing"
)

func TestValidateEnumValue(t *testing.T) {
	allowed := []string{"enabled", "disabled", "not_set"}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty value is valid", "", false},
		{"matches first option", "enabled", false},
		{"matches middle option", "disabled", false},
		{"matches last option", "not_set", false},
		{"mismatch returns error", "bogus", true},
		{"case sensitive mismatch", "Enabled", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnumValue("some-flag", tt.value, allowed)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateEnumValue(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
			if err != nil {
				msg := err.Error()
				if !strings.Contains(msg, "some-flag") || !strings.Contains(msg, tt.value) {
					t.Errorf("error message missing flag name or value: %q", msg)
				}
				for _, a := range allowed {
					if !strings.Contains(msg, a) {
						t.Errorf("error message missing allowed value %q: %q", a, msg)
					}
				}
			}
		})
	}
}

func TestParseBoolStringFlag(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantNil bool
		wantVal bool
		wantErr bool
	}{
		{"empty returns nil", "", true, false, false},
		{"true", "true", false, true, false},
		{"false", "false", false, false, false},
		{"invalid returns error", "yes", false, false, true},
		{"capitalized is invalid", "True", false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBoolStringFlag("flag", tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseBoolStringFlag(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil pointer, got %v", *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected non-nil pointer")
			}
			if *got != tt.wantVal {
				t.Errorf("expected %v, got %v", tt.wantVal, *got)
			}
		})
	}
}

func TestValidateConcurrency(t *testing.T) {
	tests := []struct {
		name    string
		val     int
		wantErr bool
	}{
		{"zero invalid", 0, true},
		{"negative invalid", -1, true},
		{"one valid", 1, false},
		{"middle valid", 10, false},
		{"twenty valid", 20, false},
		{"twenty-one invalid", 21, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConcurrency(tt.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConcurrency(%d) error = %v, wantErr %v", tt.val, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDelay(t *testing.T) {
	tests := []struct {
		name    string
		val     int
		wantErr bool
	}{
		{"negative invalid", -1, true},
		{"zero valid", 0, false},
		{"middle valid", 30, false},
		{"max valid", 600, false},
		{"over max invalid", 601, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDelay(tt.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDelay(%d) error = %v, wantErr %v", tt.val, err, tt.wantErr)
			}
		})
	}
}

func TestValidateConcurrencyAndDelay(t *testing.T) {
	tests := []struct {
		name        string
		concurrency int
		delay       int
		wantErr     bool
	}{
		{"both defaults valid", 1, 0, false},
		{"only concurrency set", 5, 0, false},
		{"only delay set", 1, 30, false},
		{"both set invalid", 5, 30, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConcurrencyAndDelay(tt.concurrency, tt.delay)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConcurrencyAndDelay(%d, %d) error = %v, wantErr %v", tt.concurrency, tt.delay, err, tt.wantErr)
			}
		})
	}
}
