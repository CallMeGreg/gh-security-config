package types

// Organization represents a GitHub organization
type Organization struct {
	Login string `json:"login"`
}

// MembershipStatus represents the user's membership status in an organization
type MembershipStatus struct {
	IsMember bool
	IsOwner  bool
	Role     string
}
