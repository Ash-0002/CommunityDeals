package service_test

// Unit tests for CampaignService business logic.
//
// These tests use a lightweight in-memory fake repository so they run without
// a real database — fast, no docker-compose needed.
//
//   go test ./internal/service/... -v -run TestCampaignService -count=1

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/community-platform/backend/internal/domain"
	"github.com/community-platform/backend/internal/dto"
	"github.com/community-platform/backend/internal/repository"
	"github.com/community-platform/backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Fake repositories
// ─────────────────────────────────────────────────────────────────────────────

// fakeCommunityRepo satisfies repository.CommunityRepository (only GetByID needed here).
type fakeCommunityRepo struct {
	communities map[string]*domain.Community
}

func (f *fakeCommunityRepo) GetByID(_ context.Context, id string) (*domain.Community, error) {
	c, ok := f.communities[id]
	if !ok {
		return nil, repository.ErrCommunityNotFound
	}
	return c, nil
}

// Unimplemented stubs — tests that hit these will panic with a clear message.
func (f *fakeCommunityRepo) Create(_ context.Context, c *domain.Community) error {
	panic("not implemented")
}
func (f *fakeCommunityRepo) List(_ context.Context, _ repository.CommunityFilter) ([]*domain.Community, int, error) {
	panic("not implemented")
}
func (f *fakeCommunityRepo) GetMember(_ context.Context, _, _ string) (*domain.CommunityMember, error) {
	panic("not implemented")
}
func (f *fakeCommunityRepo) AddMember(_ context.Context, _ *domain.CommunityMember) error {
	panic("not implemented")
}
func (f *fakeCommunityRepo) UpdateMemberStatus(_ context.Context, _, _ string, _ domain.MemberStatus, _ domain.MemberRole) error {
	panic("not implemented")
}
func (f *fakeCommunityRepo) ListMembers(_ context.Context, _ string, _, _ int) ([]*domain.CommunityMember, int, error) {
	panic("not implemented")
}
func (f *fakeCommunityRepo) ListByUserID(_ context.Context, _ string, _, _ int) ([]*domain.Community, int, error) {
	panic("not implemented")
}
func (f *fakeCommunityRepo) IsAdmin(_ context.Context, _, _ string) (bool, error) {
	panic("not implemented")
}
func (f *fakeCommunityRepo) AdminCount(_ context.Context, _ string) (int, error) {
	panic("not implemented")
}

// ─────────────────────────────────────────────────────────────────────────────

// fakeCampaignRepo is a thread-safe in-memory implementation of CampaignRepository.
type fakeCampaignRepo struct {
	mu           sync.Mutex
	campaigns    map[string]*domain.Campaign
	tiers        map[string][]domain.CampaignPricingTier      // campaignID → tiers
	participants map[string]map[string]*domain.CampaignParticipant // campaignID → userID → participant
}

func newFakeCampaignRepo() *fakeCampaignRepo {
	return &fakeCampaignRepo{
		campaigns:    make(map[string]*domain.Campaign),
		tiers:        make(map[string][]domain.CampaignPricingTier),
		participants: make(map[string]map[string]*domain.CampaignParticipant),
	}
}

func (f *fakeCampaignRepo) Create(_ context.Context, c *domain.Campaign, tiers []domain.CampaignPricingTier) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.campaigns[c.ID] = c
	f.tiers[c.ID] = tiers
	f.participants[c.ID] = make(map[string]*domain.CampaignParticipant)
	return nil
}

func (f *fakeCampaignRepo) GetByID(_ context.Context, id string) (*domain.Campaign, []domain.CampaignPricingTier, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.campaigns[id]
	if !ok {
		return nil, nil, repository.ErrCampaignNotFound
	}
	return c, f.tiers[id], nil
}

func (f *fakeCampaignRepo) GetBySlug(_ context.Context, slug string) (*domain.Campaign, []domain.CampaignPricingTier, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.campaigns {
		if c.Slug == slug {
			return c, f.tiers[c.ID], nil
		}
	}
	return nil, nil, repository.ErrCampaignNotFound
}

func (f *fakeCampaignRepo) List(_ context.Context, _ dto.ListCampaignsQuery) ([]*domain.Campaign, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []*domain.Campaign
	for _, c := range f.campaigns {
		out = append(out, c)
	}
	return out, len(out), nil
}

func (f *fakeCampaignRepo) JoinCampaign(_ context.Context, campaignID, userID string) (*repository.JoinResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	c, ok := f.campaigns[campaignID]
	if !ok {
		return nil, repository.ErrCampaignNotFound
	}
	if !c.Status.IsJoinable() {
		return nil, repository.ErrCampaignNotJoinable
	}
	if c.MaxParticipants != nil && c.ParticipantCount >= *c.MaxParticipants {
		return nil, repository.ErrCampaignFull
	}
	participants := f.participants[campaignID]
	if p, exists := participants[userID]; exists && p.Status == domain.ParticipantStatusActive {
		return nil, repository.ErrAlreadyJoined
	}

	newCount := c.ParticipantCount + 1
	c.ParticipantCount = newCount

	// auto-transition
	if newCount >= c.MinParticipants && c.Status == domain.CampaignStatusPublished {
		c.Status = domain.CampaignStatusMinimumReached
	}

	// price locked
	tiers := f.tiers[campaignID]
	price := calculatePrice(tiers, newCount)

	now := time.Now()
	participants[userID] = &domain.CampaignParticipant{
		CampaignID:  campaignID,
		UserID:      userID,
		Status:      domain.ParticipantStatusActive,
		PriceLocked: price,
		JoinedAt:    &now,
	}

	return &repository.JoinResult{
		NewParticipantCount: newCount,
		PriceLocked:         price,
		NewStatus:           c.Status,
	}, nil
}

func (f *fakeCampaignRepo) LeaveCampaign(_ context.Context, campaignID, userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	c, ok := f.campaigns[campaignID]
	if !ok {
		return repository.ErrCampaignNotFound
	}
	participants := f.participants[campaignID]
	p, exists := participants[userID]
	if !exists || p.Status != domain.ParticipantStatusActive {
		return repository.ErrNotJoined
	}
	p.Status = domain.ParticipantStatusLeft
	if c.ParticipantCount > 0 {
		c.ParticipantCount--
	}
	if c.ParticipantCount < c.MinParticipants && c.Status == domain.CampaignStatusMinimumReached {
		c.Status = domain.CampaignStatusPublished
	}
	return nil
}

func (f *fakeCampaignRepo) ListParticipants(_ context.Context, campaignID string, _, _ int) ([]*domain.CampaignParticipant, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []*domain.CampaignParticipant
	for _, p := range f.participants[campaignID] {
		out = append(out, p)
	}
	return out, len(out), nil
}

func (f *fakeCampaignRepo) UpdateStatus(_ context.Context, id string, _ domain.CampaignStatus) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.campaigns[id]; !ok {
		return repository.ErrCampaignNotFound
	}
	return nil
}

func (f *fakeCampaignRepo) IsParticipant(_ context.Context, campaignID, userID string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.participants[campaignID][userID]
	return ok && p.Status == domain.ParticipantStatusActive, nil
}

// calculatePrice mirrors the real repo helper.
func calculatePrice(tiers []domain.CampaignPricingTier, count int) int64 {
	if len(tiers) == 0 {
		return 0
	}
	var price int64 = tiers[0].Price
	for _, t := range tiers {
		if count >= t.MinCount {
			price = t.Price
		}
	}
	return price
}

// ─────────────────────────────────────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────────────────────────────────────

func newTestCommunity(id string) *domain.Community {
	return &domain.Community{
		ID:     id,
		Name:   "Test Society",
		Type:   domain.CommunityTypeSociety,
		Status: domain.CommunityStatusActive,
	}
}

func setupSvc(communityID string) (service.CampaignService, *fakeCampaignRepo) {
	fakeComm := &fakeCommunityRepo{
		communities: map[string]*domain.Community{
			communityID: newTestCommunity(communityID),
		},
	}
	fakeCamp := newFakeCampaignRepo()
	svc := service.NewCampaignService(fakeCamp, fakeComm, "")
	return svc, fakeCamp
}

func createCampaignReq(communityID string, minP, maxP int) dto.CreateCampaignRequest {
	future := time.Now().UTC().Add(7 * 24 * time.Hour).Format(time.RFC3339)
	req := dto.CreateCampaignRequest{
		CommunityID:     communityID,
		Title:           "Car Wash Drive",
		Description:     "Monthly car wash for the society",
		ServiceType:     "car_wash",
		MinParticipants: minP,
		CampaignEndDate: future,
		PricingTiers: []dto.PricingTierInput{
			{MinCount: 1, Price: 60000, Label: "Standard"},
			{MinCount: 10, Price: 50000, Label: "Group"},
			{MinCount: 25, Price: 40000, Label: "Bulk"},
		},
	}
	if maxP > 0 {
		req.MaxParticipants = &maxP
	}
	return req
}

// ─────────────────────────────────────────────────────────────────────────────
// Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestCampaignService_Create_ValidRequest(t *testing.T) {
	svc, _ := setupSvc("comm1")
	ctx := context.Background()

	resp, err := svc.Create(ctx, "user1", createCampaignReq("comm1", 5, 50))
	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.NotEmpty(t, resp.Slug)
	assert.Equal(t, "PUBLISHED", resp.Status)
	assert.Equal(t, 3, len(resp.PricingTiers))
}

func TestCampaignService_Create_InvalidCommunity(t *testing.T) {
	svc, _ := setupSvc("comm1")
	ctx := context.Background()

	_, err := svc.Create(ctx, "user1", createCampaignReq("nonexistent", 5, 50))
	assert.ErrorIs(t, err, service.ErrCommunityNotFound)
}

func TestCampaignService_Create_InvalidTiers_PriceNotDecreasing(t *testing.T) {
	svc, _ := setupSvc("comm1")
	ctx := context.Background()

	future := time.Now().UTC().Add(7 * 24 * time.Hour).Format(time.RFC3339)
	req := dto.CreateCampaignRequest{
		CommunityID:     "comm1",
		Title:           "Bad Tiers",
		ServiceType:     "car_wash",
		MinParticipants: 1,
		CampaignEndDate: future,
		PricingTiers: []dto.PricingTierInput{
			{MinCount: 1, Price: 40000},
			{MinCount: 10, Price: 60000}, // price increases — invalid
		},
	}
	_, err := svc.Create(ctx, "user1", req)
	assert.ErrorIs(t, err, service.ErrTiersNotValid)
}

func TestCampaignService_Create_InvalidTiers_FirstTierMustStartAtOne(t *testing.T) {
	svc, _ := setupSvc("comm1")
	ctx := context.Background()

	future := time.Now().UTC().Add(7 * 24 * time.Hour).Format(time.RFC3339)
	req := dto.CreateCampaignRequest{
		CommunityID:     "comm1",
		Title:           "Bad Tiers",
		ServiceType:     "car_wash",
		MinParticipants: 1,
		CampaignEndDate: future,
		PricingTiers: []dto.PricingTierInput{
			{MinCount: 5, Price: 60000}, // first tier doesn't start at 1
		},
	}
	_, err := svc.Create(ctx, "user1", req)
	assert.ErrorIs(t, err, service.ErrTiersNotValid)
}

func TestCampaignService_Join_CurrentPriceInResponse(t *testing.T) {
	svc, repo := setupSvc("comm1")
	ctx := context.Background()

	resp, err := svc.Create(ctx, "user1", createCampaignReq("comm1", 5, 50))
	require.NoError(t, err)

	joinResp, err := svc.Join(ctx, resp.ID, "userA")
	require.NoError(t, err)
	// First joiner: tier 1 = 60000 paise
	assert.Equal(t, int64(60000), joinResp.PriceLocked)
	assert.Contains(t, joinResp.Message, "joined")

	// Verify in fake repo
	camp, _, _ := repo.GetByID(ctx, resp.ID)
	assert.Equal(t, 1, camp.ParticipantCount)
}

func TestCampaignService_Join_AlreadyJoinedError(t *testing.T) {
	svc, _ := setupSvc("comm1")
	ctx := context.Background()

	resp, _ := svc.Create(ctx, "user1", createCampaignReq("comm1", 5, 50))
	_, err := svc.Join(ctx, resp.ID, "userA")
	require.NoError(t, err)

	_, err = svc.Join(ctx, resp.ID, "userA")
	assert.ErrorIs(t, err, service.ErrAlreadyJoined)
}

func TestCampaignService_Join_NextTierInfoInResponse(t *testing.T) {
	svc, _ := setupSvc("comm1")
	ctx := context.Background()

	resp, _ := svc.Create(ctx, "user1", createCampaignReq("comm1", 5, 50))

	// Join 9 users — all still in tier 1, next tier at 10
	for i := 0; i < 9; i++ {
		userID := fmt.Sprintf("user%d", i)
		joinResp, err := svc.Join(ctx, resp.ID, userID)
		require.NoError(t, err)
		if i < 8 {
			// Not yet at tier boundary
			assert.NotNil(t, joinResp.NextTier, "expected next tier info for join %d", i+1)
		}
	}
}

func TestCampaignService_Leave_NotJoinedError(t *testing.T) {
	svc, _ := setupSvc("comm1")
	ctx := context.Background()

	resp, _ := svc.Create(ctx, "user1", createCampaignReq("comm1", 5, 50))
	err := svc.Leave(ctx, resp.ID, "neverJoined")
	assert.ErrorIs(t, err, service.ErrNotJoined)
}

func TestCampaignService_Leave_ThenRejoin(t *testing.T) {
	svc, repo := setupSvc("comm1")
	ctx := context.Background()

	resp, _ := svc.Create(ctx, "user1", createCampaignReq("comm1", 5, 50))

	_, err := svc.Join(ctx, resp.ID, "userA")
	require.NoError(t, err)

	err = svc.Leave(ctx, resp.ID, "userA")
	require.NoError(t, err)

	_, err = svc.Join(ctx, resp.ID, "userA")
	require.NoError(t, err, "re-join after leaving must succeed")

	camp, _, _ := repo.GetByID(ctx, resp.ID)
	assert.Equal(t, 1, camp.ParticipantCount)
}

func TestCampaignService_UpdateStatus_InvalidTransition(t *testing.T) {
	svc, _ := setupSvc("comm1")
	ctx := context.Background()

	resp, _ := svc.Create(ctx, "user1", createCampaignReq("comm1", 5, 50))
	// PUBLISHED → COMPLETED is not a valid transition
	err := svc.UpdateStatus(ctx, resp.ID, "user1", domain.CampaignStatusCompleted)
	assert.ErrorIs(t, err, service.ErrInvalidStatusChange)
}

func TestCampaignService_GetBySlug_NotFound(t *testing.T) {
	svc, _ := setupSvc("comm1")
	ctx := context.Background()

	_, err := svc.GetBySlug(ctx, "no-such-slug", "")
	assert.ErrorIs(t, err, service.ErrCampaignNotFound)
}

// TestCampaignService_PricingTier_BreakPoints verifies that the service
// returns the correct CurrentPrice at known count breakpoints.
func TestCampaignService_PricingTier_BreakPoints(t *testing.T) {
	svc, _ := setupSvc("comm1")
	ctx := context.Background()

	resp, _ := svc.Create(ctx, "user1", createCampaignReq("comm1", 5, 100))

	tests := []struct {
		joinN         int    // total joins up to this point
		wantPricePaise int64
	}{
		{1, 60000},  // tier 1
		{9, 60000},  // still tier 1
		{10, 50000}, // tier 2 kicks in
		{24, 50000}, // still tier 2
		{25, 40000}, // tier 3
	}

	joinedSoFar := 0
	for _, tt := range tests {
		for joinedSoFar < tt.joinN {
			userID := fmt.Sprintf("joinuser%d", joinedSoFar)
			joinResp, err := svc.Join(ctx, resp.ID, userID)
			require.NoError(t, err, "join %d failed", joinedSoFar+1)
			joinedSoFar++
			if joinedSoFar == tt.joinN {
				assert.Equal(t, tt.wantPricePaise, joinResp.PriceLocked,
					"join #%d: wrong price locked", joinedSoFar)
			}
		}
	}
}
