package dto

// ── Request DTOs ──────────────────────────────────────────────────────────────

// PricingTierInput defines one pricing bracket when creating a campaign.
type PricingTierInput struct {
	MinCount int   `json:"min_count" binding:"required,min=1"`
	MaxCount int   `json:"max_count" binding:"min=0"` // 0 = unlimited / last tier
	Price    int64 `json:"price"     binding:"required,min=1"` // in paise
}

// CreateCampaignRequest is the body for POST /campaigns.
type CreateCampaignRequest struct {
	CommunityID     string             `json:"community_id"     binding:"required"`
	ServiceName     string             `json:"service_name"     binding:"required,min=2,max=100"`
	Title           string             `json:"title"            binding:"required,min=2,max=200"`
	Description     string             `json:"description"      binding:"omitempty,max=1000"`
	MinParticipants int                `json:"min_participants" binding:"required,min=1"`
	MaxParticipants int                `json:"max_participants" binding:"min=0"` // 0 = no cap
	ServiceDate     string             `json:"service_date"     binding:"required"` // RFC3339
	StartDate       string             `json:"start_date"       binding:"required"` // RFC3339
	EndDate         string             `json:"end_date"         binding:"required"` // RFC3339
	PricingTiers    []PricingTierInput `json:"pricing_tiers"    binding:"required,min=1,dive"`
}

// UpdateCampaignStatusRequest is the body for PATCH /campaigns/:id/status.
type UpdateCampaignStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// ListCampaignsQuery are URL query params for GET /campaigns and GET /communities/:id/campaigns.
type ListCampaignsQuery struct {
	CommunityID string `form:"community_id"`
	Status      string `form:"status"`
	Page        int    `form:"page,default=1"`
	Limit       int    `form:"limit,default=20"`
}

// ── Response DTOs ─────────────────────────────────────────────────────────────

// PricingTierResponse is a safe view of a pricing tier.
type PricingTierResponse struct {
	MinCount int   `json:"min_count"`
	MaxCount int   `json:"max_count"`
	Price    int64 `json:"price"`      // in paise
	TierOrder int  `json:"tier_order"`
}

// CampaignResponse is a full, safe view of a Campaign.
// This is what the campaign detail page consumes.
type CampaignResponse struct {
	ID               string                `json:"id"`
	CommunityID      string                `json:"community_id"`
	VendorID         string                `json:"vendor_id"`
	ServiceName      string                `json:"service_name"`
	Title            string                `json:"title"`
	Description      string                `json:"description"`
	ImageURL         string                `json:"image_url,omitempty"`
	MinParticipants  int                   `json:"min_participants"`
	MaxParticipants  int                   `json:"max_participants"`
	ServiceDate      string                `json:"service_date"`
	StartDate        string                `json:"start_date"`
	EndDate          string                `json:"end_date"`
	Status           string                `json:"status"`
	ParticipantCount int                   `json:"participant_count"`
	Slug             string                `json:"slug"`
	ShareURL         string                `json:"share_url"` // constructed by the handler
	PricingTiers     []PricingTierResponse `json:"pricing_tiers"`

	// Computed helpers — the frontend uses these directly
	CurrentPrice     int64                 `json:"current_price"`      // price right now in paise
	NextTier         *PricingTierResponse  `json:"next_tier,omitempty"` // nil if at best tier
	ParticipantsToNextTier int             `json:"participants_to_next_tier"` // how many more to unlock next price

	// Is the requesting user already joined?
	IsJoined bool `json:"is_joined"`

	CreatedAt string `json:"created_at"`
}

// CampaignListItem is a lighter view used in list endpoints.
type CampaignListItem struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	ServiceName      string `json:"service_name"`
	Status           string `json:"status"`
	ParticipantCount int    `json:"participant_count"`
	MinParticipants  int    `json:"min_participants"`
	CurrentPrice     int64  `json:"current_price"`
	EndDate          string `json:"end_date"`
	ServiceDate      string `json:"service_date"`
	ShareURL         string `json:"share_url"`
}

// JoinCampaignResponse is returned when a user successfully joins.
type JoinCampaignResponse struct {
	ParticipantCount int                  `json:"participant_count"` // new count after join
	PriceLocked      int64                `json:"price_locked"`      // price this user locked in
	CurrentTier      PricingTierResponse  `json:"current_tier"`
	NextTier         *PricingTierResponse `json:"next_tier,omitempty"`
	Message          string               `json:"message"`
}

// PaginatedCampaignsResponse wraps a list with pagination metadata.
type PaginatedCampaignsResponse struct {
	Campaigns []CampaignListItem `json:"campaigns"`
	Total     int                `json:"total"`
	Page      int                `json:"page"`
	Limit     int                `json:"limit"`
	HasMore   bool               `json:"has_more"`
}
