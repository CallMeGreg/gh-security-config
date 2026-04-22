package types

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestConfigurationExistsError_Message(t *testing.T) {
	err := &ConfigurationExistsError{ConfigName: "cfg", OrgName: "my-org"}
	msg := err.Error()
	if !strings.Contains(msg, "cfg") || !strings.Contains(msg, "my-org") {
		t.Errorf("error message missing fields: %q", msg)
	}
}

func TestConfigurationExistsError_ErrorsAs(t *testing.T) {
	base := &ConfigurationExistsError{ConfigName: "cfg", OrgName: "my-org"}
	wrapped := fmt.Errorf("context: %w", base)

	var target *ConfigurationExistsError
	if !errors.As(wrapped, &target) {
		t.Fatal("errors.As should unwrap ConfigurationExistsError")
	}
	if target.ConfigName != "cfg" || target.OrgName != "my-org" {
		t.Errorf("unexpected fields: %+v", target)
	}
}

func TestDependabotUnavailableError_Message(t *testing.T) {
	err := &DependabotUnavailableError{Feature: "alerts", OrgName: "my-org"}
	msg := err.Error()
	if !strings.Contains(msg, "alerts") || !strings.Contains(msg, "my-org") {
		t.Errorf("error message missing fields: %q", msg)
	}
}

func TestDependabotUnavailableError_ErrorsAs(t *testing.T) {
	base := &DependabotUnavailableError{Feature: "security-updates", OrgName: "my-org"}
	wrapped := fmt.Errorf("outer: %w", base)

	var target *DependabotUnavailableError
	if !errors.As(wrapped, &target) {
		t.Fatal("errors.As should unwrap DependabotUnavailableError")
	}
	if target.Feature != "security-updates" {
		t.Errorf("unexpected feature: %q", target.Feature)
	}
}
