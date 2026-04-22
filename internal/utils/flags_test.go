package utils

import (
	"path/filepath"
	"testing"
)

func TestHasOrgTargeting(t *testing.T) {
	tests := []struct {
		name  string
		flags *CommonFlags
		want  bool
	}{
		{"none set", &CommonFlags{}, false},
		{"org set", &CommonFlags{Org: "x"}, true},
		{"org-list set", &CommonFlags{OrgListPath: "x.csv"}, true},
		{"all-orgs set", &CommonFlags{AllOrgs: true}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasOrgTargeting(tt.flags); got != tt.want {
				t.Errorf("HasOrgTargeting = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateOrgFlags_RequiresOneTarget(t *testing.T) {
	err := ValidateOrgFlags(&CommonFlags{})
	if err == nil {
		t.Fatal("expected error when no targeting flag is set")
	}
}

func TestValidateOrgFlags_AcceptsSingleOrg(t *testing.T) {
	if err := ValidateOrgFlags(&CommonFlags{Org: "my-org"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateOrgFlags_AcceptsAllOrgs(t *testing.T) {
	if err := ValidateOrgFlags(&CommonFlags{AllOrgs: true}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateOrgFlags_RejectsInvalidOrgName(t *testing.T) {
	if err := ValidateOrgFlags(&CommonFlags{Org: "bad name"}); err == nil {
		t.Error("expected error for org name with space")
	}
	if err := ValidateOrgFlags(&CommonFlags{Org: "bad/name"}); err == nil {
		t.Error("expected error for org name with slash")
	}
}

func TestValidateOrgFlags_MissingCSV(t *testing.T) {
	err := ValidateOrgFlags(&CommonFlags{OrgListPath: filepath.Join(t.TempDir(), "nope.csv")})
	if err == nil {
		t.Fatal("expected error for missing CSV")
	}
}

func TestValidateOrgFlags_EmptyCSV(t *testing.T) {
	path := writeTempCSV(t, "\n   \n")
	err := ValidateOrgFlags(&CommonFlags{OrgListPath: path})
	if err == nil {
		t.Fatal("expected error for CSV with no valid organizations")
	}
}

func TestValidateOrgFlags_ValidCSV(t *testing.T) {
	path := writeTempCSV(t, "org-one\norg-two\n")
	if err := ValidateOrgFlags(&CommonFlags{OrgListPath: path}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateOrgFlagsOptional_AllowsEmpty(t *testing.T) {
	if err := ValidateOrgFlagsOptional(&CommonFlags{}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateOrgFlagsOptional_RejectsInvalidOrg(t *testing.T) {
	if err := ValidateOrgFlagsOptional(&CommonFlags{Org: "bad name"}); err == nil {
		t.Error("expected error for org name with space")
	}
}

func TestValidateOrgFlagsOptional_RejectsMissingCSV(t *testing.T) {
	err := ValidateOrgFlagsOptional(&CommonFlags{OrgListPath: filepath.Join(t.TempDir(), "nope.csv")})
	if err == nil {
		t.Fatal("expected error for missing CSV")
	}
}
