package domain

import (
	"database/sql"
	"time"
)

// CampaignStatus represents the full lifecycle of a campaign.
// Transitions are strictly enforced in the service layer.
type CampaignStatus string

const (
	// DRAFT — created by vendor/admin but not yet visible to community members.
	CampaignStatusDraft CampaignStatus = "DRAFT"

	// PUBLISHED — visible to the community; members can join.
	CampaignStatusPublished CampaignStatus = "PUBLISHED"

	// MINIMUM_REACHED — participant count crossed the minimum threshold.
	// Pricing is locked at the current tier. Campaign can still accept more participants.
	CampaignStatusMinimumReached CampaignStatus = "MINIMUM_REACHED"

	// CONFIRMED — admin/vendor manually confirmed the campaign will proceed.
	// No new participants allowed. Payment collection begins.
	CampaignStatusConfirmed CampaignStatus = "CONFIRMED"

	// PAYMENT_PENDING — payment links/reminders sent to all participants.
	CampaignStatusPaymentPending CampaignStatus = "PAYMENT_PENDING"

	// PAYMENT_COMPLETED — all participants have paid.
	CampaignStatusPaymentCompleted CampaignStatus = "PAYMENT_COMPLETED"

	// IN_PROGRESS — service is being delivered.
	CampaignStatusInProgress CampaignStatus = "IN_PROGRESS"

	// COMPLETED — service delivered successfully.
	CampaignStatusCompleted CampaignStatus = "COMPLETED"

	// CANCELLED — cancelled by vendor or admin before service.
	CampaignStatusCancelled CampaignStatus = "CANCELLED"

	// EXPIRED — deadline passed without reaching minimum participants.
	CampaignStatusExpired CampaignStatus = "EXPIRED"
)

// IsJoinable returns true when new participants can still join.
func (s CampaignStatus) IsJoinable() bool {
	return s == CampaignStatusPublished || s == CampaignStatusMinimumReached
}

// IsActive returns true when the campaign is running (not terminal).
func (s CampaignStatus) IsActive() bool {
	switch s {
	case CampaignStatusCancelled, CampaignStatusExpired, CampaignStatusCompleted:
		return false
	}
	return true
}

// ValidTransitions defines the allowed next statuses from each status.
// The service layer uses this to enforce the state machine.
var ValidTransitions = map[CampaignStatus][]CampaignStatus{
	CampaignStatusDraft:            {CampaignStatusPublished, CampaignStatusCancelled},
	CampaignStatusPublished:        {CampaignStatusMinimumReached, CampaignStatusExpired, CampaignStatusCancelled},
	CampaignStatusMinimumReached:   {CampaignStatusConfirmed, CampaignStatusExpired, CampaignStatusCancelled},
	CampaignStatusConfirmed:        {CampaignStatusPaymentPending, CampaignStatusCancelled},
	CampaignStatusPaymentPending:   {CampaignStatusPaymentCompleted, CampaignStatusCancelled},
	CampaignStatusPaymentCompleted: {CampaignStatusInProgress},
	CampaignStatusInProgress:       {CampaignStatusCompleted, CampaignStatusCancelled},
	CampaignStatusCompleted:        {},
	CampaignStatusCancelled:        {},
	CampaignStatusExpired:          {},
}

// CanTransitionTo returns true if moving to `next` is a valid state transition.
func (s CampaignStatus) CanTransitionTo(next CampaignStatus) bool {
	for _, allowed := range ValidTransitions[s] {
		if allowed == next {
			return true
		}
	}
	return false
}

// ParticipantStatus tracks an individual's joining state.
type ParticipantStatus string

const (
	ParticipantStatusJoined    ParticipantStatus = "JOINED"
	ParticipantStatusLeft      ParticipantStatus = "LEFT"
	ParticipantStatusCancelled ParticipantStatus = "CANCELLED"
)

// Campaign is the core deal/offer entity.
// Example: "Car Washing This Sunday — Green Valley Society"
type Campaign struct {
	ID          string         `db:"id"`
	CommunityID string         `db:"community_id"`
	VendorID    string         `db:"vendor_id"`    // the user who created it (vendor/admin)
	ServiceName string         `db:"service_name"` // e.g. "Car Washing"
	Title       string         `db:"title"`        // e.g. "Car Washing This Sunday"
	Description string         `db:"description"`
	ImageURL    string         `db:"image_url"`

	// Participant limits
	MinParticipants int `db:"min_participants"` // minimum to confirm the campaign
	MaxParticipants int `db:"max_participants"` // 0 = no limit

	// Scheduling
	ServiceDate  time.Time  `db:"service_date"`  // when the service will be delivered
	StartDate    time.Time  `db:"start_date"`    // when users can start joining
	EndDate      time.Time  `db:"end_date"`      // deadline for joining

	Status CampaignStatus `db:"status"`

	// Denormalized for fast reads (updated atomically with participant inserts)
	ParticipantCount int `db:"participant_count"`

	// Shareable link slug — e.g. "car-wash-gv-aug30"
	Slug string `db:"slug"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// IsExpired returns true when the campaign deadline has passed.
func (c *Campaign) IsExpired() bool {
	return time.Now().After(c.EndDate)
}

// IsFull returns true when max participants has been reached.
func (c *Campaign) IsFull() bool {
	return c.MaxParticipants > 0 && c.ParticipantCount >= c.MaxParticipants
}

// CampaignPricingTier defines a price bracket based on participant count.
// Example: 10–24 participants → ₹500
type CampaignPricingTier struct {
	ID         string `db:"id"`
	CampaignID string `db:"campaign_id"`
	MinCount   int    `db:"min_count"` // inclusive lower bound
	MaxCount   int    `db:"max_count"` // inclusive upper bound (0 = unlimited)
	Price      int64  `db:"price"`     // price in paise (₹500 = 50000 paise) — use integers, never floats for money
	TierOrder  int    `db:"tier_order"` // display/sort order, starting from 1

	CreatedAt time.Time `db:"created_at"`
}

// CampaignParticipant records a user's membership in a campaign.
type CampaignParticipant struct {
	ID         string            `db:"id"`
	CampaignID string            `db:"campaign_id"`
	UserID     string            `db:"user_id"`
	Status     ParticipantStatus `db:"status"`
	// Price locked when the user joined (in paise)
	// Stored so we know what to charge even if tier changes later.
	PriceLocked int64        `db:"price_locked"`
	JoinedAt    time.Time    `db:"joined_at"`
	LeftAt      sql.NullTime `db:"left_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
}

// IsActive returns true when the participant is still in the campaign.
func (p *CampaignParticipant) IsActive() bool {
	return p.Status == ParticipantStatusJoined
}
