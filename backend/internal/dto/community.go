package dto

// ── Request DTOs ──────────────────────────────────────────────────────────────

// CreateCommunityRequest is the body for POST /communities.
type CreateCommunityRequest struct {
	Name             string `json:"name"              binding:"required,min=2,max=100"`
	Description      string `json:"description"       binding:"omitempty,max=500"`
	Type             string `json:"type"              binding:"required"`
	City             string `json:"city"              binding:"omitempty,max=100"`
	State            string `json:"state"             binding:"omitempty,max=100"`
	PinCode          string `json:"pin_code"          binding:"omitempty,max=10"`
	Address          string `json:"address"           binding:"omitempty,max=300"`
	RequiresApproval bool   `json:"requires_approval"`
}

// JoinCommunityRequest is the body for POST /communities/:id/join.
// invite_code is optional — required only when the community uses invite-code access.
type JoinCommunityRequest struct {
	InviteCode string `json:"invite_code" binding:"omitempty"`
}

// ListCommunitiesQuery are URL query params for GET /communities.
type ListCommunitiesQuery struct {
	Type    string `form:"type"`
	City    string `form:"city"`
	PinCode string `form:"pin_code"`
	Page    int    `form:"page,default=1"`
	Limit   int    `form:"limit,default=20"`
}

// ── Response DTOs ─────────────────────────────────────────────────────────────

// CommunityResponse is a safe, serialisable view of a Community.
type CommunityResponse struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Type             string `json:"type"`
	Status           string `json:"status"`
	City             string `json:"city,omitempty"`
	State            string `json:"state,omitempty"`
	PinCode          string `json:"pin_code,omitempty"`
	Address          string `json:"address,omitempty"`
	RequiresApproval bool   `json:"requires_approval"`
	InviteCode       string `json:"invite_code,omitempty"` // only shown to admins
	LogoURL          string `json:"logo_url,omitempty"`
	MemberCount      int    `json:"member_count"`
	CreatedByID      string `json:"created_by_id"`
	CreatedAt        string `json:"created_at"`
}

// CommunityMemberResponse is a safe view of a CommunityMember.
type CommunityMemberResponse struct {
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	JoinedAt  string `json:"joined_at,omitempty"`
}

// JoinCommunityResponse confirms a join request.
type JoinCommunityResponse struct {
	MembershipStatus string `json:"membership_status"` // PENDING or APPROVED
	Message          string `json:"message"`
}

// PaginatedCommunitiesResponse wraps a list with pagination metadata.
type PaginatedCommunitiesResponse struct {
	Communities []CommunityResponse `json:"communities"`
	Total       int                 `json:"total"`
	Page        int                 `json:"page"`
	Limit       int                 `json:"limit"`
	HasMore     bool                `json:"has_more"`
}
