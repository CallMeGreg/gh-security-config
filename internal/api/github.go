package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2"
	"github.com/pterm/pterm"

	"github.com/callmegreg/gh-security-config/internal/types"
)

// GetCurrentUser returns the current GitHub user login
func GetCurrentUser() (string, error) {
	userResponse, _, err := gh.Exec("api", "user", "-q", ".login")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(userResponse.String()), nil
}

// CheckSingleOrganizationMembership checks if the current user has access to an organization
func CheckSingleOrganizationMembership(org string) (types.MembershipStatus, error) {
	// Get current user's login first
	currentUser, err := GetCurrentUser()
	if err != nil {
		return types.MembershipStatus{}, fmt.Errorf("failed to get current user: %w", err)
	}

	// Use REST API to check membership and role directly
	userResponse, stderr, err := gh.Exec("api", "-H", "Accept: application/vnd.github+json", "-H", "X-GitHub-Api-Version: 2022-11-28", fmt.Sprintf("/orgs/%s/memberships/%s", org, currentUser))
	if err != nil {
		// If we get a 404 or similar error, the user is likely not a member
		if strings.Contains(stderr.String(), "404") || strings.Contains(stderr.String(), "Not Found") {
			return types.MembershipStatus{IsMember: false, IsOwner: false, Role: "none"}, nil
		}
		return types.MembershipStatus{IsMember: false, IsOwner: false, Role: "none"}, nil
	}

	var membership struct {
		State string `json:"state"`
		Role  string `json:"role"`
	}

	if err := json.Unmarshal(userResponse.Bytes(), &membership); err != nil {
		pterm.Warning.Printf("Failed to parse membership data for organization '%s': %v\n", org, err)
		return types.MembershipStatus{IsMember: false, IsOwner: false, Role: "none"}, nil
	}

	// Check if membership is active and determine role
	if membership.State == "active" {
		isOwner := membership.Role == "admin"
		return types.MembershipStatus{
			IsMember: true,
			IsOwner:  isOwner,
			Role:     membership.Role,
		}, nil
	}

	// If state is not active, treat as not a member
	return types.MembershipStatus{IsMember: false, IsOwner: false, Role: "none"}, nil
}

// ValidateMembershipAndSkip is a helper function that checks membership and returns appropriate ProcessingResult
func ValidateMembershipAndSkip(org string) *types.ProcessingResult {
	status, err := CheckSingleOrganizationMembership(org)
	if err != nil {
		pterm.Warning.Printf("Failed to check membership for organization '%s': %v, skipping\n", org, err)
		return &types.ProcessingResult{Organization: org, Skipped: true}
	}
	if !status.IsMember {
		pterm.Warning.Printf("Skipping organization '%s': You are not a member\n", org)
		return &types.ProcessingResult{Organization: org, Skipped: true}
	}
	if !status.IsOwner {
		pterm.Warning.Printf("Skipping organization '%s': You are a member but not an owner\n", org)
		return &types.ProcessingResult{Organization: org, Skipped: true}
	}
	return nil // No skip needed
}
