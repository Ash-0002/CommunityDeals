package repository_test

// Integration tests for CampaignRepository.
//
// These tests require a live PostgreSQL database.  The easiest way to run them:
//
//   docker-compose up -d postgres          # start DB
//   make migrate-up                        # apply migrations 000001–000003
//   go test ./internal/repository/... -v -run TestCampaign -count=1
//
// Set POSTGRES_DSN in your environment (or .env) to override the default:
//   export POSTGRES_DSN="postgres://cp_user:cp_password@localhost:5432/community_platform?sslmode=disable"
//
// The tests create their own isolated data (random IDs) so they are safe to
// run against a shared development database.  They clean up after themselves.

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/community-platform/backend/internal/domain"
	"github.com/community-platform/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// ─────────────────────────────────────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────────────────────────────────────

func testDSN() string {
	if dsn := os.Getenv("POSTGRES_DSN"); dsn != "" {
		return dsn
	}
	return "postgres://cp_user:cp_password@localhost:5432/community_platform?sslmode=disable"
}

func mustConnect(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Connect("postgres", testDSN())
	require.NoError(t, err, "failed to connect to test database — is docker-compose running and migrate-up done?")
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db
}

func uid() string {
	// Tiny unique ID helper — good enough for test fixtures.
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// insertTestUser inserts a minimal user row and returns the ID.
func insertTestUser(t *testing.T, db *sqlx.DB) string {
	t.Helper()
	id := "u_" + uid()
	_, err := db.Exec(`
		INSERT INTO users (id, phone, name, role, status, is_verified)
		VALUES ($1, $2, 'Test User', 'MEMBER', 'ACTIVE', true)`,
		id, "+91"+uid()[:10],
	)
	require.NoError(t, err)
	t.Cleanup(func() { db.Exec(`DELETE FROM users WHERE id=$1`, id) }) // nolint
	return id
}

// insertTestCommunity inserts a minimal community row and returns the ID.
func insertTestCommunity(t *testing.T, db *sqlx.DB, creatorID string) string {
	t.Helper()
	id := "c_" + uid()
	_, err := db.Exec(`
		INSERT INTO communities (id, name, type, status, created_by_id, invite_code, requires_approval)
		VALUES ($1, 'Test Community', 'SOCIETY', 'ACTIVE', $2, $3, false)`,
		id, creatorID, "INV"+uid()[:4],
	)
	require.NoError(t, err)
	t.Cleanup(func() { db.Exec(`DELETE FROM communities WHERE id=$1`, id) }) // nolint
	return id
}

// insertTestCampaign inserts a campaign and pricing tiers; returns the campaign ID.
func insertTestCampaign(
	t *testing.T,
	db *sqlx.DB,
	communityID, creatorID string,
	minP, maxP int,
	tiers []domain.CampaignPricingTier,
) string {
	t.Helper()
	id := "camp_" + uid()
	slug := "test-campaign-" + uid()

	_, err := db.Exec(`
		INSERT INTO campaigns
			(id, community_id, created_by_id, title, description, service_type,
			 slug, status, min_participants, max_participants, participant_count)
		VALUES ($1,$2,$3,'Test Campaign','','car_wash',$4,'PUBLISHED',$5,$6,0)`,
		id, communityID, creatorID, slug, minP, maxP,
	)
	require.NoError(t, err)

	for _, tier := range tiers {
		_, err = db.Exec(`
			INSERT INTO campaign_pricing_tiers
				(id, campaign_id, min_count, price, label, tier_order)
			VALUES ($1,$2,$3,$4,$5,$6)`,
			"tier_"+uid(), id, tier.MinCount, tier.Price, tier.Label, tier.TierOrder,
		)
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		db.Exec(`DELETE FROM campaign_participants WHERE campaign_id=$1`, id) // nolint
		db.Exec(`DELETE FROM campaign_pricing_tiers WHERE campaign_id=$1`, id) // nolint
		db.Exec(`DELETE FROM campaigns WHERE id=$1`, id)                       // nolint
	})
	return id
}

// ─────────────────────────────────────────────────────────────────────────────
// test suite
// ─────────────────────────────────────────────────────────────────────────────

type CampaignRepositoryTestSuite struct {
	suite.Suite
	db          *sqlx.DB
	repo        repository.CampaignRepository
	creatorID   string
	communityID string
}

func (s *CampaignRepositoryTestSuite) SetupSuite() {
	s.db = mustConnect(s.T())
	s.repo = repository.NewCampaignRepository(s.db)
	s.creatorID = insertTestUser(s.T(), s.db)
	s.communityID = insertTestCommunity(s.T(), s.db, s.creatorID)
}

func (s *CampaignRepositoryTestSuite) TearDownSuite() {
	s.db.Close()
}

func TestCampaignRepositorySuite(t *testing.T) {
	suite.Run(t, new(CampaignRepositoryTestSuite))
}

// ─────────────────────────────────────────────────────────────────────────────
// T1 – basic join / leave round-trip
// ─────────────────────────────────────────────────────────────────────────────

func (s *CampaignRepositoryTestSuite) TestJoinAndLeave_Basic() {
	ctx := context.Background()
	userID := insertTestUser(s.T(), s.db)

	campaignID := insertTestCampaign(s.T(), s.db, s.communityID, s.creatorID, 1, 100,
		[]domain.CampaignPricingTier{
			{MinCount: 1, Price: 60000, TierOrder: 1},
		})

	// Join
	result, err := s.repo.JoinCampaign(ctx, campaignID, userID)
	s.Require().NoError(err)
	s.Equal(1, result.NewParticipantCount)
	s.Equal(int64(60000), result.PriceLocked)

	// Participant count in DB
	var count int
	s.Require().NoError(s.db.QueryRow(`SELECT participant_count FROM campaigns WHERE id=$1`, campaignID).Scan(&count))
	s.Equal(1, count)

	// Leave
	err = s.repo.LeaveCampaign(ctx, campaignID, userID)
	s.Require().NoError(err)

	s.Require().NoError(s.db.QueryRow(`SELECT participant_count FROM campaigns WHERE id=$1`, campaignID).Scan(&count))
	s.Equal(0, count, "participant_count must decrement after leave")
}

// ─────────────────────────────────────────────────────────────────────────────
// T2 – duplicate join returns ErrAlreadyJoined
// ─────────────────────────────────────────────────────────────────────────────

func (s *CampaignRepositoryTestSuite) TestJoin_DuplicateReturnsError() {
	ctx := context.Background()
	userID := insertTestUser(s.T(), s.db)
	campaignID := insertTestCampaign(s.T(), s.db, s.communityID, s.creatorID, 1, 50,
		[]domain.CampaignPricingTier{{MinCount: 1, Price: 50000, TierOrder: 1}})

	_, err := s.repo.JoinCampaign(ctx, campaignID, userID)
	s.Require().NoError(err)

	_, err = s.repo.JoinCampaign(ctx, campaignID, userID)
	s.ErrorIs(err, repository.ErrAlreadyJoined, "second join must return ErrAlreadyJoined")
}

// ─────────────────────────────────────────────────────────────────────────────
// T3 – max_participants cap is never exceeded (concurrent joins)
// ─────────────────────────────────────────────────────────────────────────────
//
// Strategy: spin up N goroutines all calling JoinCampaign simultaneously.
// Only (max_participants) should succeed; the rest must return ErrCampaignFull.
// After the dust settles:
//   • participant_count == max_participants  (never over)
//   • ACTIVE rows in campaign_participants == max_participants

func (s *CampaignRepositoryTestSuite) TestJoin_ConcurrentDoesNotExceedMax() {
	const total = 30
	const maxP = 10

	ctx := context.Background()
	campaignID := insertTestCampaign(s.T(), s.db, s.communityID, s.creatorID, 1, maxP,
		[]domain.CampaignPricingTier{{MinCount: 1, Price: 60000, TierOrder: 1}})

	// Create total distinct users
	users := make([]string, total)
	for i := range users {
		users[i] = insertTestUser(s.T(), s.db)
	}

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		joined   []string
		fullErrs int32
	)

	for _, uid := range users {
		wg.Add(1)
		go func(userID string) {
			defer wg.Done()
			res, err := s.repo.JoinCampaign(ctx, campaignID, userID)
			if err == nil {
				mu.Lock()
				joined = append(joined, userID)
				_ = res
				mu.Unlock()
			} else if err == repository.ErrCampaignFull {
				atomic.AddInt32(&fullErrs, 1)
			}
			// ErrAlreadyJoined should not appear here (each user is unique)
		}(uid)
	}
	wg.Wait()

	s.Len(joined, maxP, "exactly max_participants goroutines should succeed")
	s.EqualValues(total-maxP, atomic.LoadInt32(&fullErrs), "remaining goroutines should see ErrCampaignFull")

	// DB must agree
	var dbCount int
	s.Require().NoError(s.db.QueryRow(`SELECT participant_count FROM campaigns WHERE id=$1`, campaignID).Scan(&dbCount))
	s.Equal(maxP, dbCount, "participant_count in DB must equal max_participants")

	var activeRows int
	s.Require().NoError(s.db.QueryRow(
		`SELECT COUNT(*) FROM campaign_participants WHERE campaign_id=$1 AND status='ACTIVE'`, campaignID,
	).Scan(&activeRows))
	s.Equal(maxP, activeRows, "ACTIVE participant rows must equal max_participants")
}

// ─────────────────────────────────────────────────────────────────────────────
// T4 – pricing tier transitions at the right counts
// ─────────────────────────────────────────────────────────────────────────────
//
// Tiers:  1–9 → ₹600 (60000p), 10–24 → ₹500 (50000p), 25+ → ₹400 (40000p)
// The N-th joiner should have their price_locked set to the tier that applies
// at count N.

func (s *CampaignRepositoryTestSuite) TestJoin_PricingTierLocking() {
	ctx := context.Background()
	campaignID := insertTestCampaign(s.T(), s.db, s.communityID, s.creatorID, 1, 30,
		[]domain.CampaignPricingTier{
			{MinCount: 1, Price: 60000, TierOrder: 1},
			{MinCount: 10, Price: 50000, TierOrder: 2},
			{MinCount: 25, Price: 40000, TierOrder: 3},
		})

	// Helper: join one user and return price_locked
	join := func(n int) int64 {
		uid := insertTestUser(s.T(), s.db)
		res, err := s.repo.JoinCampaign(ctx, campaignID, uid)
		s.Require().NoError(err, "join #%d should succeed", n)
		return res.PriceLocked
	}

	// Joins 1–9: tier 1 → 60000
	for i := 1; i <= 9; i++ {
		s.Equal(int64(60000), join(i), "join %d: expected ₹600 tier", i)
	}

	// Join 10: tier 2 → 50000
	s.Equal(int64(50000), join(10), "join 10: expected ₹500 tier (10+ group)")

	// Joins 11–24: still tier 2
	for i := 11; i <= 24; i++ {
		s.Equal(int64(50000), join(i), "join %d: expected ₹500 tier", i)
	}

	// Join 25: tier 3 → 40000
	s.Equal(int64(40000), join(25), "join 25: expected ₹400 tier (25+ group)")
}

// ─────────────────────────────────────────────────────────────────────────────
// T5 – auto-transition: PUBLISHED → MINIMUM_REACHED when minParticipants hit
// ─────────────────────────────────────────────────────────────────────────────

func (s *CampaignRepositoryTestSuite) TestJoin_AutoTransitionToMinimumReached() {
	ctx := context.Background()
	campaignID := insertTestCampaign(s.T(), s.db, s.communityID, s.creatorID, 5, 50,
		[]domain.CampaignPricingTier{{MinCount: 1, Price: 60000, TierOrder: 1}})

	for i := 1; i <= 4; i++ {
		uid := insertTestUser(s.T(), s.db)
		_, err := s.repo.JoinCampaign(ctx, campaignID, uid)
		s.Require().NoError(err)
	}

	// Status should still be PUBLISHED after 4 joins (min = 5)
	var status string
	s.Require().NoError(s.db.QueryRow(`SELECT status FROM campaigns WHERE id=$1`, campaignID).Scan(&status))
	s.Equal("PUBLISHED", status, "status should still be PUBLISHED before min is hit")

	// 5th join crosses min_participants
	uid := insertTestUser(s.T(), s.db)
	_, err := s.repo.JoinCampaign(ctx, campaignID, uid)
	s.Require().NoError(err)

	s.Require().NoError(s.db.QueryRow(`SELECT status FROM campaigns WHERE id=$1`, campaignID).Scan(&status))
	s.Equal("MINIMUM_REACHED", status, "status must auto-transition to MINIMUM_REACHED at 5th join")
}

// ─────────────────────────────────────────────────────────────────────────────
// T6 – auto-revert: MINIMUM_REACHED → PUBLISHED when count drops below min
// ─────────────────────────────────────────────────────────────────────────────

func (s *CampaignRepositoryTestSuite) TestLeave_AutoRevertToPublished() {
	ctx := context.Background()
	campaignID := insertTestCampaign(s.T(), s.db, s.communityID, s.creatorID, 3, 50,
		[]domain.CampaignPricingTier{{MinCount: 1, Price: 60000, TierOrder: 1}})

	users := make([]string, 3)
	for i := range users {
		users[i] = insertTestUser(s.T(), s.db)
		_, err := s.repo.JoinCampaign(ctx, campaignID, users[i])
		s.Require().NoError(err)
	}

	var status string
	s.Require().NoError(s.db.QueryRow(`SELECT status FROM campaigns WHERE id=$1`, campaignID).Scan(&status))
	s.Equal("MINIMUM_REACHED", status, "pre-condition: should be MINIMUM_REACHED after 3 joins")

	// One person leaves → count drops to 2 (below min=3)
	err := s.repo.LeaveCampaign(ctx, campaignID, users[0])
	s.Require().NoError(err)

	s.Require().NoError(s.db.QueryRow(`SELECT status FROM campaigns WHERE id=$1`, campaignID).Scan(&status))
	s.Equal("PUBLISHED", status, "status must revert to PUBLISHED when count drops below min")
}

// ─────────────────────────────────────────────────────────────────────────────
// T7 – cannot join a campaign that is not in a joinable state
// ─────────────────────────────────────────────────────────────────────────────

func (s *CampaignRepositoryTestSuite) TestJoin_RejectsNonJoinableStatus() {
	ctx := context.Background()
	campaignID := insertTestCampaign(s.T(), s.db, s.communityID, s.creatorID, 1, 50,
		[]domain.CampaignPricingTier{{MinCount: 1, Price: 60000, TierOrder: 1}})

	// Force status to CANCELLED
	_, err := s.db.Exec(`UPDATE campaigns SET status='CANCELLED' WHERE id=$1`, campaignID)
	s.Require().NoError(err)

	uid := insertTestUser(s.T(), s.db)
	_, err = s.repo.JoinCampaign(ctx, campaignID, uid)
	s.ErrorIs(err, repository.ErrCampaignNotJoinable, "joining a CANCELLED campaign must fail")
}

// ─────────────────────────────────────────────────────────────────────────────
// T8 – cannot join an expired campaign (campaign_end_date in the past)
// ─────────────────────────────────────────────────────────────────────────────

func (s *CampaignRepositoryTestSuite) TestJoin_RejectsExpiredCampaign() {
	ctx := context.Background()
	campaignID := insertTestCampaign(s.T(), s.db, s.communityID, s.creatorID, 1, 50,
		[]domain.CampaignPricingTier{{MinCount: 1, Price: 60000, TierOrder: 1}})

	// Set end date to yesterday
	yesterday := time.Now().UTC().Add(-24 * time.Hour)
	_, err := s.db.Exec(`UPDATE campaigns SET campaign_end_date=$1 WHERE id=$2`, yesterday, campaignID)
	s.Require().NoError(err)

	uid := insertTestUser(s.T(), s.db)
	_, err = s.repo.JoinCampaign(ctx, campaignID, uid)
	s.ErrorIs(err, repository.ErrCampaignExpired, "joining an expired campaign must fail")
}

// ─────────────────────────────────────────────────────────────────────────────
// T9 – participant_count never goes below zero on concurrent leaves
// ─────────────────────────────────────────────────────────────────────────────

func (s *CampaignRepositoryTestSuite) TestLeave_CountNeverBelowZero() {
	const n = 5
	ctx := context.Background()
	campaignID := insertTestCampaign(s.T(), s.db, s.communityID, s.creatorID, 1, 20,
		[]domain.CampaignPricingTier{{MinCount: 1, Price: 60000, TierOrder: 1}})

	// n users join
	users := make([]string, n)
	for i := range users {
		users[i] = insertTestUser(s.T(), s.db)
		_, err := s.repo.JoinCampaign(ctx, campaignID, users[i])
		s.Require().NoError(err)
	}

	// All n users leave concurrently
	var wg sync.WaitGroup
	for _, uid := range users {
		wg.Add(1)
		go func(userID string) {
			defer wg.Done()
			_ = s.repo.LeaveCampaign(ctx, campaignID, userID)
		}(uid)
	}
	wg.Wait()

	var count int
	s.Require().NoError(s.db.QueryRow(`SELECT participant_count FROM campaigns WHERE id=$1`, campaignID).Scan(&count))
	s.GreaterOrEqual(count, 0, "participant_count must never go below 0")
}

// ─────────────────────────────────────────────────────────────────────────────
// T10 – join after leave (re-join) succeeds
// ─────────────────────────────────────────────────────────────────────────────

func (s *CampaignRepositoryTestSuite) TestJoin_ReJoinAfterLeave() {
	ctx := context.Background()
	userID := insertTestUser(s.T(), s.db)
	campaignID := insertTestCampaign(s.T(), s.db, s.communityID, s.creatorID, 1, 50,
		[]domain.CampaignPricingTier{{MinCount: 1, Price: 60000, TierOrder: 1}})

	// Join, leave, join again
	_, err := s.repo.JoinCampaign(ctx, campaignID, userID)
	s.Require().NoError(err)

	err = s.repo.LeaveCampaign(ctx, campaignID, userID)
	s.Require().NoError(err)

	_, err = s.repo.JoinCampaign(ctx, campaignID, userID)
	s.Require().NoError(err, "re-join after leaving must succeed")

	var count int
	s.Require().NoError(s.db.QueryRow(`SELECT participant_count FROM campaigns WHERE id=$1`, campaignID).Scan(&count))
	s.Equal(1, count)

	var activeRows int
	s.Require().NoError(s.db.QueryRow(
		`SELECT COUNT(*) FROM campaign_participants WHERE campaign_id=$1 AND user_id=$2 AND status='ACTIVE'`,
		campaignID, userID,
	).Scan(&activeRows))
	s.Equal(1, activeRows, "only one ACTIVE row per user per campaign")
}
