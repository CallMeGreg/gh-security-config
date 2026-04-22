package utils

import (
	"strings"
	"testing"
)

// TestBuildReplicationCommand_DefaultsAreOmitted ensures that default values
// for concurrency and delay don't end up in the replication command.
func TestBuildReplicationCommand_DefaultsAreOmitted(t *testing.T) {
	got := BuildReplicationCommand("generate", map[string]interface{}{
		"enterprise-slug": "e",
		"all-orgs":        true,
		"concurrency":     1,
		"delay":           0,
	})
	if strings.Contains(got, "--concurrency") {
		t.Errorf("default concurrency should not be emitted: %s", got)
	}
	if strings.Contains(got, "--delay") {
		t.Errorf("default delay should not be emitted: %s", got)
	}
}

// TestBuildReplicationCommand_FalseBoolsOmitted ensures false bool flags are not emitted.
func TestBuildReplicationCommand_FalseBoolsOmitted(t *testing.T) {
	got := BuildReplicationCommand("generate", map[string]interface{}{
		"enterprise-slug": "e",
		"all-orgs":        false,
	})
	if strings.Contains(got, "--all-orgs") {
		t.Errorf("false bool flag should not be emitted: %s", got)
	}
}

// TestBuildReplicationCommand_EmptyStringsOmitted ensures empty string flags are skipped.
func TestBuildReplicationCommand_EmptyStringsOmitted(t *testing.T) {
	got := BuildReplicationCommand("generate", map[string]interface{}{
		"enterprise-slug": "e",
		"all-orgs":        true,
		"config-name":     "",
	})
	if strings.Contains(got, "--config-name") {
		t.Errorf("empty string flag should not be emitted: %s", got)
	}
}

// TestBuildReplicationCommand_NilValuesOmitted ensures nil values are skipped.
func TestBuildReplicationCommand_NilValuesOmitted(t *testing.T) {
	got := BuildReplicationCommand("generate", map[string]interface{}{
		"enterprise-slug": "e",
		"all-orgs":        true,
		"config-name":     nil,
	})
	if strings.Contains(got, "--config-name") {
		t.Errorf("nil flag should not be emitted: %s", got)
	}
}

// TestBuildReplicationCommand_UnknownKeysIgnored ensures flags not in flagOrder are skipped.
func TestBuildReplicationCommand_UnknownKeysIgnored(t *testing.T) {
	got := BuildReplicationCommand("generate", map[string]interface{}{
		"enterprise-slug": "e",
		"all-orgs":        true,
		"not-a-real-flag": "value",
	})
	if strings.Contains(got, "not-a-real-flag") {
		t.Errorf("unknown flag should be ignored: %s", got)
	}
}

// TestBuildReplicationCommand_DeterministicOrder ensures flag order is stable.
func TestBuildReplicationCommand_DeterministicOrder(t *testing.T) {
	flags := map[string]interface{}{
		"enterprise-slug": "e",
		"template-org":    "t",
		"org":             "o",
		"scope":           "all",
	}
	got := BuildReplicationCommand("apply", flags)
	// Expected: enterprise-slug before template-org before org before scope.
	idxEnt := strings.Index(got, "--enterprise-slug")
	idxTmpl := strings.Index(got, "--template-org")
	idxOrg := strings.Index(got, "--org ")
	idxScope := strings.Index(got, "--scope")
	if !(idxEnt < idxTmpl && idxTmpl < idxOrg && idxOrg < idxScope) {
		t.Errorf("flags not in expected order: %s", got)
	}
}
