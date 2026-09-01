package domain

import "time"

// CommunityType classifies the kind of community.
// The platform is designed generically — not only for housing societies.
type CommunityType string

const (
	CommunityTypeSociety             CommunityType = "SOCIETY"
	CommunityTypeApartment           CommunityType = "APARTMENT"
	CommunityTypeResidentialComplex  CommunityType = "RESIDENTIAL_COMPLEX"
	CommunityTypeOffice              CommunityType = "OFFICE"
	CommunityTypeCollege             CommunityType = "COLLEGE"
	CommunityTypeCorporateGroup      CommunityType = "CORPORATE_GROUP"
	CommunityTypePrivateGroup        CommunityType = "PRIVATE_GROUP"
	CommunityTypePartnerCommunity    CommunityType = "PARTNER_COMMUNITY"
)

// CommunityStatus tracks whether the community is usable.
type CommunityStatus string

const (
	CommunityStatusActive   CommunityStatus = "ACTIVE"
	CommunityStatusInactive CommunityStatus = "INACTIVE"
	CommunityStatusArchived CommunityStatus = "ARCHIVED"
)

// MemberRole is the role a user holds within a specific community.
type MemberRole string

const (
	MemberRoleMember MemberRole = "MEMBER"
	MemberRoleAdmin  MemberRole = "ADMIN"
)

// MemberStatus tracks membership state.
type MemberStatus string

const (
	MemberStatusPending  MemberStatus = "PENDING"  // awaiting admin approval
	MemberStatusApproved MemberStatus = "APPROVED" // active member
	MemberStatusRejected MemberStatus = "REJECTED"
	MemberStatusRemoved  MemberStatus = "REMOVED"
)

// Community is the core grouping entity. Campaigns belong to communities.
type Community struct {
	ID          string          `db:"id"`
	Name        string          `db:"name"`
	Description string          `db:"description"`
	Type        CommunityType   `db:"type"`
	Status      CommunityStatus `db:"status"`
	// Location fields (optional — helps vendors target the right area)
	City        string `db:"city"`
	State       string `db:"state"`
	PinCode     string `db:"pin_code"`
	Address     string `db:"address"`
	// Settings
	RequiresApproval bool   `db:"requires_approval"` // if true, join requests need admin approval
	InviteCode       string `db:"invite_code"`       // short code for easy joining
	LogoURL          string `db:"logo_url"`
	// Ownership
	CreatedByID string    `db:"created_by_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// IsActive checks if the community accepts new members and campaigns.
func (c *Community) IsActive() bool {
	return c.Status == CommunityStatusActive
}

// CommunityMember represents a user's membership in a community.
type CommunityMember struct {
	ID          string       `db:"id"`
	CommunityID string       `db:"community_id"`
	UserID      string       `db:"user_id"`
	Role        MemberRole   `db:"role"`
	Status      MemberStatus `db:"status"`
	JoinedAt    *time.Time   `db:"joined_at"`  // nil until approved
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
}

// IsActive returns true when the membership is approved and usable.
func (m *CommunityMember) IsActive() bool {
	return m.Status == MemberStatusApproved
}

// IsAdmin returns true when the member has admin rights in the community.
func (m *CommunityMember) IsAdmin() bool {
	return m.Role == MemberRoleAdmin && m.Status == MemberStatusApproved
}
