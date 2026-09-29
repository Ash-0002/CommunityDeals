package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/community-platform/backend/internal/domain"
	"github.com/community-platform/backend/internal/dto"
	"github.com/community-platform/backend/internal/repository"
)

// Sentinel errors for the campaign flow.
var (
	ErrCampaignNotFound     = errors.New("campaign not found")
	ErrCampaignNotJoinable  = errors.New("campaign is not open for joining")
	ErrCampaignFull         = errors.New("campaign is full")
	ErrAlreadyJoined        = errors.New("you have already joined this campaign")
	ErrNotJoined            = errors.New("you have not joined this campaign")
	ErrInvalidStatusChange  = errors.New("invalid status transition")
	ErrTiersNotValid        = errors.New("pricing tiers are invalid")
)

// CampaignService handles campaign business logic.
type CampaignService interface {
	Create(ctx context.Context, creatorID string, req dto.CreateCampaignRequest) (*dto.CampaignResponse, error)
	GetByID(ctx context.Context, id, requestingUserID string) (*dto.CampaignResponse, error)
	GetBySlug(ctx context.Context, slug, requestingUserID string) (*dto.CampaignResponse, error)
	List(ctx context.Context, filter dto.ListCampaignsQuery, requestingUserID string) (*dto.PaginatedCampaignsResponse, error)
	Join(ctx context.Context, campaignID, userID string) (*dto.JoinCampaignResponse, error)
	Leave(ctx context.Context, campaignID, userID string) error
	UpdateStatus(ctx context.Context, campaignID, requestingUserID string, newStatus domain.CampaignStatus) error
	ListParticipants(ctx context.Context, campaignID string, page, limit int) ([]*domain.CampaignParticipant, int, error)
}

type campaignService struct {
	campaignRepo  repository.CampaignRepository
	communityRepo repository.CommunityRepository
	appBaseURL    string
}

// NewCampaignService creates a CampaignService with all dependencies injected.
// appBaseURL is used to build shareable campaign links, e.g. "https://communitydeals.app".
func NewCampaignService(campaignRepo repository.CampaignRepository, communityRepo repository.CommunityRepository, appBaseURL string) CampaignService {
	if appBaseURL == "" {
		appBaseURL = "http://localhost:3000"
	}
	return &campaignService{
		campaignRepo:  campaignRepo,
		communityRepo: communityRepo,
		appBaseURL:    strings.TrimSuffix(appBaseURL, "/"),
	}
}

// Create validates and persists a new campaign with its pricing tiers.
func (s *campaignService) Create(ctx context.Context, creatorID string, req dto.CreateCampaignRequest) (*dto.CampaignResponse, error) {
	// Validate community exists
	community, err := s.communityRepo.FindByID(ctx, req.CommunityID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCommunityNotFound
		}
		return nil, fmt.Errorf("finding community: %w", err)
	}
	if !community.IsActive() {
		return nil, ErrCommunityInactive
	}

	// Parse dates
	serviceDate, err := time.Parse(time.RFC3339, req.ServiceDate)
	if err != nil {
		return nil, fmt.Errorf("invalid service_date format (use RFC3339): %w", err)
	}
	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format: %w", err)
	}
	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format: %w", err)
	}

	if endDate.Before(startDate) {
		return nil, fmt.Errorf("end_date must be after start_date")
	}
	if serviceDate.Before(endDate) {
		return nil, fmt.Errorf("service_date must be on or after end_date")
	}

	// Validate pricing tiers
	if err := validatePricingTiers(req.PricingTiers, req.MinParticipants); err != nil {
		return nil, err
	}

	// Build campaign
	now := time.Now()
	slug := generateSlug(req.Title)

	campaign := &domain.Campaign{
		ID:               newUUID(),
		CommunityID:      req.CommunityID,
		VendorID:         creatorID,
		ServiceName:      strings.TrimSpace(req.ServiceName),
		Title:            strings.TrimSpace(req.Title),
		Description:      strings.TrimSpace(req.Description),
		ImageURL:         strings.TrimSpace(req.ImageURL),
		MinParticipants:  req.MinParticipants,
		MaxParticipants:  req.MaxParticipants,
		ServiceDate:      serviceDate,
		StartDate:        startDate,
		EndDate:          endDate,
		Status:           domain.CampaignStatusPublished, // publish immediately in MVP
		ParticipantCount: 0,
		Slug:             slug,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.campaignRepo.Create(ctx, campaign); err != nil {
		return nil, fmt.Errorf("creating campaign: %w", err)
	}

	// Build and persist pricing tiers
	tiers := make([]*domain.CampaignPricingTier, len(req.PricingTiers))
	for i, t := range req.PricingTiers {
		tiers[i] = &domain.CampaignPricingTier{
			ID:         newUUID(),
			CampaignID: campaign.ID,
			MinCount:   t.MinCount,
			MaxCount:   t.MaxCount,
			Price:      t.Price,
			TierOrder:  i + 1,
		}
	}
	if err := s.campaignRepo.CreatePricingTiers(ctx, tiers); err != nil {
		return nil, fmt.Errorf("creating pricing tiers: %w", err)
	}

	return s.buildCampaignResponse(ctx, campaign, tiers, false), nil
}

// GetByID fetches a campaign by ID, with pricing tiers and join status.
func (s *campaignService) GetByID(ctx context.Context, id, requestingUserID string) (*dto.CampaignResponse, error) {
	campaign, err := s.campaignRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCampaignNotFound
		}
		return nil, fmt.Errorf("getting campaign: %w", err)
	}

	tiers, err := s.campaignRepo.GetPricingTiers(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting pricing tiers: %w", err)
	}

	isJoined := false
	if requestingUserID != "" {
		isJoined, _ = s.campaignRepo.IsParticipant(ctx, id, requestingUserID)
	}

	return s.buildCampaignResponse(ctx, campaign, tiers, isJoined), nil
}

// GetBySlug fetches a campaign by its shareable slug.
func (s *campaignService) GetBySlug(ctx context.Context, slug, requestingUserID string) (*dto.CampaignResponse, error) {
	campaign, err := s.campaignRepo.FindBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCampaignNotFound
		}
		return nil, err
	}
	return s.GetByID(ctx, campaign.ID, requestingUserID)
}

// List returns a paginated list of campaigns with optional filters.
func (s *campaignService) List(ctx context.Context, q dto.ListCampaignsQuery, requestingUserID string) (*dto.PaginatedCampaignsResponse, error) {
	filter := repository.CampaignFilter{
		CommunityID: q.CommunityID,
		Status:      q.Status,
		Page:        q.Page,
		Limit:       q.Limit,
	}

	campaigns, total, err := s.campaignRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing campaigns: %w", err)
	}

	items := make([]dto.CampaignListItem, 0, len(campaigns))
	for _, c := range campaigns {
		tiers, _ := s.campaignRepo.GetPricingTiers(ctx, c.ID)
		currentPrice := calculateCurrentPrice(tiers, c.ParticipantCount)
		isJoined := false
		if requestingUserID != "" {
			isJoined, _ = s.campaignRepo.IsParticipant(ctx, c.ID, requestingUserID)
		}
		firstTierPrice := currentPrice
		if len(tiers) > 0 {
			firstTierPrice = tiers[0].Price
		}
		items = append(items, dto.CampaignListItem{
			ID:               c.ID,
			Title:            c.Title,
			ServiceName:      c.ServiceName,
			ImageURL:         c.ImageURL,
			Status:           string(c.Status),
			ParticipantCount: c.ParticipantCount,
			MinParticipants:  c.MinParticipants,
			CurrentPrice:     currentPrice,
			EndDate:          c.EndDate.Format(time.RFC3339),
			ServiceDate:      c.ServiceDate.Format(time.RFC3339),
			ShareURL:         s.buildShareURL(c.Slug),
			IsJoined:         isJoined,
			FirstTierPrice:   firstTierPrice,
		})
	}

	return &dto.PaginatedCampaignsResponse{
		Campaigns: items,
		Total:     total,
		Page:      q.Page,
		Limit:     q.Limit,
		HasMore:   (q.Page * q.Limit) < total,
	}, nil
}

// Join safely adds a user to a campaign using a DB-level transaction with SELECT FOR UPDATE.
func (s *campaignService) Join(ctx context.Context, campaignID, userID string) (*dto.JoinCampaignResponse, error) {
	participant, newCount, err := s.campaignRepo.JoinCampaign(ctx, campaignID, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return nil, ErrCampaignNotFound
		case errors.Is(err, repository.ErrAlreadyJoined):
			return nil, ErrAlreadyJoined
		case errors.Is(err, repository.ErrCampaignFull):
			return nil, ErrCampaignFull
		default:
			return nil, fmt.Errorf("joining campaign: %w", err)
		}
	}

	tiers, err := s.campaignRepo.GetPricingTiers(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("getting tiers after join: %w", err)
	}

	currentTier := findCurrentTier(tiers, newCount)
	nextTier := findNextTier(tiers, newCount)

	resp := &dto.JoinCampaignResponse{
		ParticipantCount: newCount,
		PriceLocked:      participant.PriceLocked,
		Message:          fmt.Sprintf("You joined! %d people in so far.", newCount),
	}

	if currentTier != nil {
		resp.CurrentTier = toTierResponse(currentTier)
	}
	if nextTier != nil {
		nt := toTierResponse(nextTier)
		resp.NextTier = &nt
		needed := nextTier.MinCount - newCount
		if needed > 0 {
			resp.Message = fmt.Sprintf(
				"You joined! %d people in. %d more to unlock ₹%.0f pricing.",
				newCount, needed, float64(nextTier.Price)/100,
			)
		}
	}

	return resp, nil
}

// Leave removes a user from a campaign.
func (s *campaignService) Leave(ctx context.Context, campaignID, userID string) error {
	_, err := s.campaignRepo.LeaveCampaign(ctx, campaignID, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return ErrCampaignNotFound
		case errors.Is(err, repository.ErrNotJoined):
			return ErrNotJoined
		default:
			return fmt.Errorf("leaving campaign: %w", err)
		}
	}
	return nil
}

// UpdateStatus transitions a campaign to a new status, enforcing the state machine.
func (s *campaignService) UpdateStatus(ctx context.Context, campaignID, requestingUserID string, newStatus domain.CampaignStatus) error {
	campaign, err := s.campaignRepo.FindByID(ctx, campaignID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrCampaignNotFound
		}
		return err
	}

	if !campaign.Status.CanTransitionTo(newStatus) {
		return fmt.Errorf("%w: %s → %s", ErrInvalidStatusChange, campaign.Status, newStatus)
	}

	return s.campaignRepo.UpdateStatus(ctx, campaignID, newStatus)
}

// ListParticipants returns active participants of a campaign.
func (s *campaignService) ListParticipants(ctx context.Context, campaignID string, page, limit int) ([]*domain.CampaignParticipant, int, error) {
	return s.campaignRepo.ListParticipants(ctx, campaignID, page, limit)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// buildCampaignResponse assembles the full CampaignResponse DTO.
func (s *campaignService) buildCampaignResponse(
	ctx context.Context,
	c *domain.Campaign,
	tiers []*domain.CampaignPricingTier,
	isJoined bool,
) *dto.CampaignResponse {
	tierResponses := make([]dto.PricingTierResponse, len(tiers))
	for i, t := range tiers {
		tierResponses[i] = toTierResponse(t)
	}

	currentPrice := calculateCurrentPrice(tiers, c.ParticipantCount)
	nextTier := findNextTier(tiers, c.ParticipantCount)

	resp := &dto.CampaignResponse{
		ID:               c.ID,
		CommunityID:      c.CommunityID,
		VendorID:         c.VendorID,
		ServiceName:      c.ServiceName,
		Title:            c.Title,
		Description:      c.Description,
		ImageURL:         c.ImageURL,
		MinParticipants:  c.MinParticipants,
		MaxParticipants:  c.MaxParticipants,
		ServiceDate:      c.ServiceDate.Format(time.RFC3339),
		StartDate:        c.StartDate.Format(time.RFC3339),
		EndDate:          c.EndDate.Format(time.RFC3339),
		Status:           string(c.Status),
		ParticipantCount: c.ParticipantCount,
		Slug:             c.Slug,
		ShareURL:         s.buildShareURL(c.Slug),
		PricingTiers:     tierResponses,
		CurrentPrice:     currentPrice,
		IsJoined:         isJoined,
		CreatedAt:        c.CreatedAt.Format(time.RFC3339),
	}

	if nextTier != nil {
		nt := toTierResponse(nextTier)
		resp.NextTier = &nt
		resp.ParticipantsToNextTier = nextTier.MinCount - c.ParticipantCount
	}

	return resp
}

// calculateCurrentPrice returns the price for the current participant count.
func calculateCurrentPrice(tiers []*domain.CampaignPricingTier, count int) int64 {
	if len(tiers) == 0 {
		return 0
	}
	for _, t := range tiers {
		if count >= t.MinCount && (t.MaxCount == 0 || count <= t.MaxCount) {
			return t.Price
		}
	}
	// Below the first bracket (a brand-new campaign sits at 0 participants,
	// and tier 1 starts at 1) nothing matches. That must fall back to the
	// starting price, not tiers[last] — otherwise an empty campaign advertises
	// the deepest group discount nobody has earned yet.
	if count < tiers[0].MinCount {
		return tiers[0].Price
	}
	return tiers[len(tiers)-1].Price
}

// findCurrentTier returns the tier that applies at the given count.
func findCurrentTier(tiers []*domain.CampaignPricingTier, count int) *domain.CampaignPricingTier {
	for _, t := range tiers {
		if count >= t.MinCount && (t.MaxCount == 0 || count <= t.MaxCount) {
			return t
		}
	}
	if len(tiers) > 0 {
		return tiers[len(tiers)-1]
	}
	return nil
}

// findNextTier returns the next cheaper tier above the current count.
func findNextTier(tiers []*domain.CampaignPricingTier, count int) *domain.CampaignPricingTier {
	current := calculateCurrentPrice(tiers, count)
	for _, t := range tiers {
		// Only a genuinely cheaper bracket is worth chasing. Without the price
		// check, an empty campaign reports "1 more to unlock <the price it is
		// already showing>".
		if t.MinCount > count && t.Price < current {
			return t
		}
	}
	return nil
}

func toTierResponse(t *domain.CampaignPricingTier) dto.PricingTierResponse {
	return dto.PricingTierResponse{
		MinCount:  t.MinCount,
		MaxCount:  t.MaxCount,
		Price:     t.Price,
		TierOrder: t.TierOrder,
	}
}

// buildShareURL constructs the public shareable link for a campaign.
// Uses the configured APP_BASE_URL (web app origin) and the `/c/:slug` route.
func (s *campaignService) buildShareURL(slug string) string {
	return fmt.Sprintf("%s/c/%s", s.appBaseURL, slug)
}

// generateSlug builds a URL-safe slug from a title, appending a short random suffix.
// e.g. "Car Washing This Sunday" → "car-washing-this-sunday-a3f2"
func generateSlug(title string) string {
	// Lowercase and replace non-alphanumeric with hyphens
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug := re.ReplaceAllString(strings.Map(func(r rune) rune {
		return unicode.ToLower(r)
	}, title), "-")
	slug = strings.Trim(slug, "-")
	if len(slug) > 50 {
		slug = slug[:50]
	}
	// Append short random suffix to avoid collisions
	suffix, _ := randomHex(2)
	return slug + "-" + suffix
}

// validatePricingTiers checks that tiers are ordered and cover the minimum participant count.
func validatePricingTiers(tiers []dto.PricingTierInput, minParticipants int) error {
	if len(tiers) == 0 {
		return fmt.Errorf("%w: at least one pricing tier required", ErrTiersNotValid)
	}

	// Tiers must be ordered by MinCount ascending
	for i := 1; i < len(tiers); i++ {
		if tiers[i].MinCount <= tiers[i-1].MinCount {
			return fmt.Errorf("%w: tiers must have increasing min_count values", ErrTiersNotValid)
		}
		// Each tier's price should be lower or equal (bulk discount logic)
		if tiers[i].Price > tiers[i-1].Price {
			return fmt.Errorf("%w: higher tiers should have equal or lower prices (bulk discount)", ErrTiersNotValid)
		}
	}

	// First tier must start at 1
	if tiers[0].MinCount != 1 {
		return fmt.Errorf("%w: first tier min_count must be 1", ErrTiersNotValid)
	}

	return nil
}
