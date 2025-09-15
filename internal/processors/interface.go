package processors

import "github.com/callmegreg/gh-security-config/internal/types"

// OrganizationProcessor defines the interface for processing organizations
type OrganizationProcessor interface {
	ProcessOrganization(org string) types.ProcessingResult
}
