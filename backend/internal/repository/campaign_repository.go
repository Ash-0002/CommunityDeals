package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/community-platform/backend/internal/domain"
	"github.com/jmoiron/sqlx"
)

var (
	// ErrCampaignFull is returned when the campaign has reached max participants.
	ErrCampaignFull = errors.New("campaign is full")
	// ErrAlreadyJoined is returned when a user tries to join a campaign they are already in.
	ErrAlreadyJoined = errors.New("user has already joined this campaign")
	// ErrNotJoined is returned when a leave is attempted but user is not a participant.
	ErrNotJoined = errors.New("user has not joined this campaign")
	// ErrInvalidTransition is returned when a status change violates the state machine.
	ErrInvalidTransition = errors.New("invalid campaign status transition")
)

// CampaignRepository defines data-access operations for campaigns.
type CampaignRepository interface {
	// Campaign CRUD
	Create(ctx context.Context, c *domain.Campaign) error
	FindByID(ctx context.Context, id string) (*domain.Campaign, error)
	FindBySlug(ctx context.Context, slug string) (*domain.Campaign, error)
	List(ctx context.Context, filter CampaignFilter) ([]*domain.Campaign, int, error)
	UpdateStatus(ctx context.Context, campaignID string, status domain.CampaignStatus) error

	// Pricing tiers
	CreatePricingTiers(ctx context.Context, tiers []*domain.CampaignPricingTier) error
	GetPricingTiers(ctx context.Context, campaignID string) ([]*domain.CampaignPricingTier, error)

	// Participants — the JOIN logic uses SELECT FOR UPDATE to prevent race conditions
	JoinCampaign(ctx context.Context, campaignID, userID string) (*domain.CampaignParticipant, int, error)
	LeaveCampaign(ctx context.Context, campaignID, userID string) (int, error)
	FindParticipant(ctx context.Context, campaignID, userID string) (*domain.CampaignParticipant, error)
	ListParticipants(ctx context.Context, campaignID string, page, limit int) ([]*domain.CampaignParticipant, int, error)
	IsParticipant(ctx context.Context, campaignID, userID string) (bool, error)
}

// CampaignFilter filters the campaign list query.
type CampaignFilter struct {
	CommunityID string
	Status      string
	Page        int
	Limit       int
}

type campaignRepository struct {
	db *sqlx.DB
}

// NewCampaignRepository creates a PostgreSQL-backed CampaignRepository.
func NewCampaignRepository(db *sqlx.DB) CampaignRepository {
	return &campaignRepository{db: db}
}

// ── Campaign CRUD ──────────────────────────────────────────────────────────────

func (r *campaignRepository) Create(ctx context.Context, c *domain.Campaign) error {
	query := `
		INSERT INTO campaigns
			(id, community_id, vendor_id, service_name, title, description, image_url,
			 min_participants, max_participants, service_date, start_date, end_date,
			 status, participant_count, slug, created_at, updated_at)
		VALUES
			(:id, :community_id, :vendor_id, :service_name, :title, :description, :image_url,
			 :min_participants, :max_participants, :service_date, :start_date, :end_date,
			 :status, :participant_count, :slug, :created_at, :updated_at)
	`
	if _, err := r.db.NamedExecContext(ctx, query, c); err != nil {
		return fmt.Errorf("create campaign: %w", err)
	}
	return nil
}

func (r *campaignRepository) FindByID(ctx context.Context, id string) (*domain.Campaign, error) {
	var c domain.Campaign
	if err := r.db.GetContext(ctx, &c, `SELECT * FROM campaigns WHERE id = $1`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find campaign by id: %w", err)
	}
	return &c, nil
}

func (r *campaignRepository) FindBySlug(ctx context.Context, slug string) (*domain.Campaign, error) {
	var c domain.Campaign
	if err := r.db.GetContext(ctx, &c, `SELECT * FROM campaigns WHERE slug = $1`, slug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find campaign by slug: %w", err)
	}
	return &c, nil
}

func (r *campaignRepository) List(ctx context.Context, f CampaignFilter) ([]*domain.Campaign, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	idx := 1

	if f.CommunityID != "" {
		where += fmt.Sprintf(" AND community_id = $%d", idx)
		args = append(args, f.CommunityID)
		idx++
	}
	if f.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, f.Status)
		idx++
	}

	var total int
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM campaigns "+where, args...); err != nil {
		return nil, 0, fmt.Errorf("count campaigns: %w", err)
	}

	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	offset := (f.Page - 1) * f.Limit

	query := fmt.Sprintf(
		"SELECT * FROM campaigns %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		where, idx, idx+1,
	)
	args = append(args, f.Limit, offset)

	var campaigns []*domain.Campaign
	if err := r.db.SelectContext(ctx, &campaigns, query, args...); err != nil {
		return nil, 0, fmt.Errorf("list campaigns: %w", err)
	}
	return campaigns, total, nil
}

func (r *campaignRepository) UpdateStatus(ctx context.Context, campaignID string, status domain.CampaignStatus) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE campaigns SET status = $1, updated_at = $2 WHERE id = $3`,
		status, time.Now(), campaignID,
	)
	return err
}

// ── Pricing Tiers ─────────────────────────────────────────────────────────────

func (r *campaignRepository) CreatePricingTiers(ctx context.Context, tiers []*domain.CampaignPricingTier) error {
	if len(tiers) == 0 {
		return nil
	}
	query := `
		INSERT INTO campaign_pricing_tiers (id, campaign_id, min_count, max_count, price, tier_order)
		VALUES (:id, :campaign_id, :min_count, :max_count, :price, :tier_order)
	`
	for _, t := range tiers {
		if _, err := r.db.NamedExecContext(ctx, query, t); err != nil {
			return fmt.Errorf("insert pricing tier: %w", err)
		}
	}
	return nil
}

func (r *campaignRepository) GetPricingTiers(ctx context.Context, campaignID string) ([]*domain.CampaignPricingTier, error) {
	var tiers []*domain.CampaignPricingTier
	err := r.db.SelectContext(ctx, &tiers,
		`SELECT * FROM campaign_pricing_tiers WHERE campaign_id = $1 ORDER BY tier_order ASC`,
		campaignID,
	)
	if err != nil {
		return nil, fmt.Errorf("get pricing tiers: %w", err)
	}
	return tiers, nil
}

// ── Participants — Race-Condition-Safe JOIN ────────────────────────────────────

// JoinCampaign atomically joins a user to a campaign.
//
// Safety guarantees:
//  1. Uses a serializable transaction.
//  2. SELECT FOR UPDATE locks the campaign row — concurrent joins queue up
//     instead of reading a stale participant_count.
//  3. Checks max_participants inside the lock.
//  4. Inserts participant with a UNIQUE constraint (campaign_id, user_id)
//     as a second safety net against duplicate joins.
//  5. Atomically increments participant_count on the campaigns row.
//
// Returns the new participant record and the updated participant count.
func (r *campaignRepository) JoinCampaign(ctx context.Context, campaignID, userID string) (*domain.CampaignParticipant, int, error) {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// ── Step 1: Lock the campaign row ─────────────────────────────────────────
	// SELECT FOR UPDATE prevents any other transaction from reading or modifying
	// this row until our transaction commits or rolls back.
	var campaign domain.Campaign
	if err = tx.GetContext(ctx, &campaign,
		`SELECT * FROM campaigns WHERE id = $1 FOR UPDATE`,
		campaignID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, ErrNotFound
		}
		return nil, 0, fmt.Errorf("lock campaign: %w", err)
	}

	// ── Step 2: Business rule checks (inside the lock) ────────────────────────
	if !campaign.Status.IsJoinable() {
		return nil, 0, fmt.Errorf("campaign is not open for joining (status: %s)", campaign.Status)
	}
	if campaign.IsExpired() {
		return nil, 0, fmt.Errorf("campaign deadline has passed")
	}
	if campaign.IsFull() {
		return nil, 0, ErrCampaignFull
	}

	// ── Step 3: Check for existing active participation ───────────────────────
	var existingCount int
	if err = tx.GetContext(ctx, &existingCount,
		`SELECT COUNT(*) FROM campaign_participants WHERE campaign_id = $1 AND user_id = $2 AND status = 'JOINED'`,
		campaignID, userID,
	); err != nil {
		return nil, 0, fmt.Errorf("check existing participant: %w", err)
	}
	if existingCount > 0 {
		return nil, 0, ErrAlreadyJoined
	}

	// ── Step 4: Get current pricing tiers to lock the price ───────────────────
	var tiers []*domain.CampaignPricingTier
	if err = tx.SelectContext(ctx, &tiers,
		`SELECT * FROM campaign_pricing_tiers WHERE campaign_id = $1 ORDER BY tier_order ASC`,
		campaignID,
	); err != nil {
		return nil, 0, fmt.Errorf("get pricing tiers in tx: %w", err)
	}

	// Count after this join = current + 1
	newCount := campaign.ParticipantCount + 1
	priceLocked := calculatePriceForCount(tiers, newCount)

	// ── Step 5: Insert participant ────────────────────────────────────────────
	now := time.Now()
	participantID := generateUUID()
	participant := &domain.CampaignParticipant{
		ID:          participantID,
		CampaignID:  campaignID,
		UserID:      userID,
		Status:      domain.ParticipantStatusJoined,
		PriceLocked: priceLocked,
		JoinedAt:    now,
		UpdatedAt:   now,
	}

	if _, err = tx.NamedExecContext(ctx, `
		INSERT INTO campaign_participants (id, campaign_id, user_id, status, price_locked, joined_at, updated_at)
		VALUES (:id, :campaign_id, :user_id, :status, :price_locked, :joined_at, :updated_at)
	`, participant); err != nil {
		if isUniqueViolation(err) {
			return nil, 0, ErrAlreadyJoined
		}
		return nil, 0, fmt.Errorf("insert participant: %w", err)
	}

	// ── Step 6: Atomically increment participant_count ────────────────────────
	var updatedCount int
	if err = tx.GetContext(ctx, &updatedCount, `
		UPDATE campaigns
		SET participant_count = participant_count + 1, updated_at = $1
		WHERE id = $2
		RETURNING participant_count
	`, now, campaignID); err != nil {
		return nil, 0, fmt.Errorf("increment participant count: %w", err)
	}

	// ── Step 7: Auto-transition status if minimum reached ────────────────────
	if updatedCount >= campaign.MinParticipants && campaign.Status == domain.CampaignStatusPublished {
		if _, err = tx.ExecContext(ctx,
			`UPDATE campaigns SET status = $1, updated_at = $2 WHERE id = $3`,
			domain.CampaignStatusMinimumReached, now, campaignID,
		); err != nil {
			return nil, 0, fmt.Errorf("update status to minimum_reached: %w", err)
		}
	}

	// ── Step 8: Commit ────────────────────────────────────────────────────────
	if err = tx.Commit(); err != nil {
		return nil, 0, fmt.Errorf("commit join transaction: %w", err)
	}

	return participant, updatedCount, nil
}

// LeaveCampaign marks a participant as LEFT and decrements the count atomically.
// Returns the updated participant count.
func (r *campaignRepository) LeaveCampaign(ctx context.Context, campaignID, userID string) (int, error) {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Lock campaign
	var campaign domain.Campaign
	if err = tx.GetContext(ctx, &campaign,
		`SELECT * FROM campaigns WHERE id = $1 FOR UPDATE`, campaignID,
	); err != nil {
		return 0, fmt.Errorf("lock campaign: %w", err)
	}

	// Only allow leaving if campaign is still joinable
	if !campaign.Status.IsJoinable() {
		return 0, fmt.Errorf("cannot leave a campaign with status %s", campaign.Status)
	}

	// Mark participant as LEFT
	now := time.Now()
	result, err := tx.ExecContext(ctx, `
		UPDATE campaign_participants
		SET status = 'LEFT', updated_at = $1
		WHERE campaign_id = $2 AND user_id = $3 AND status = 'JOINED'
	`, now, campaignID, userID)
	if err != nil {
		return 0, fmt.Errorf("mark participant as left: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		_ = tx.Rollback()
		return 0, ErrNotJoined
	}

	// Decrement count
	var updatedCount int
	if err = tx.GetContext(ctx, &updatedCount, `
		UPDATE campaigns
		SET participant_count = GREATEST(participant_count - 1, 0), updated_at = $1
		WHERE id = $2
		RETURNING participant_count
	`, now, campaignID); err != nil {
		return 0, fmt.Errorf("decrement participant count: %w", err)
	}

	// Revert to PUBLISHED if count drops below minimum
	if updatedCount < campaign.MinParticipants && campaign.Status == domain.CampaignStatusMinimumReached {
		if _, err = tx.ExecContext(ctx,
			`UPDATE campaigns SET status = $1, updated_at = $2 WHERE id = $3`,
			domain.CampaignStatusPublished, now, campaignID,
		); err != nil {
			return 0, fmt.Errorf("revert status to published: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit leave transaction: %w", err)
	}
	return updatedCount, nil
}

func (r *campaignRepository) FindParticipant(ctx context.Context, campaignID, userID string) (*domain.CampaignParticipant, error) {
	var p domain.CampaignParticipant
	err := r.db.GetContext(ctx, &p,
		`SELECT * FROM campaign_participants WHERE campaign_id = $1 AND user_id = $2 ORDER BY joined_at DESC LIMIT 1`,
		campaignID, userID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find participant: %w", err)
	}
	return &p, nil
}

func (r *campaignRepository) ListParticipants(ctx context.Context, campaignID string, page, limit int) ([]*domain.CampaignParticipant, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM campaign_participants WHERE campaign_id = $1 AND status = 'JOINED'`,
		campaignID,
	); err != nil {
		return nil, 0, fmt.Errorf("count participants: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var participants []*domain.CampaignParticipant
	err := r.db.SelectContext(ctx, &participants, `
		SELECT * FROM campaign_participants
		WHERE campaign_id = $1 AND status = 'JOINED'
		ORDER BY joined_at ASC
		LIMIT $2 OFFSET $3
	`, campaignID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list participants: %w", err)
	}
	return participants, total, nil
}

func (r *campaignRepository) IsParticipant(ctx context.Context, campaignID, userID string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM campaign_participants WHERE campaign_id = $1 AND user_id = $2 AND status = 'JOINED'`,
		campaignID, userID,
	)
	return count > 0, err
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// calculatePriceForCount returns the price (in paise) for the given participant count
// based on the sorted pricing tiers.
func calculatePriceForCount(tiers []*domain.CampaignPricingTier, count int) int64 {
	if len(tiers) == 0 {
		return 0
	}
	// Walk tiers in order; return the price of the tier that contains `count`.
	for _, t := range tiers {
		if count >= t.MinCount && (t.MaxCount == 0 || count <= t.MaxCount) {
			return t.Price
		}
	}
	// Fallback: use the last tier's price (open-ended)
	return tiers[len(tiers)-1].Price
}

// generateUUID returns a random UUID v4 string.
func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
