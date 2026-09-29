// Command seed populates the database with demo data so the web, mobile,
// and public-link flows can all be clicked through end-to-end without
// manually creating a community, campaigns, and users first.
//
// Usage:
//
//	cd backend
//	make docker-up && make migrate-up
//	go run ./cmd/seed
//
// Safe to re-run: users are matched by phone number and campaigns by title,
// so running it twice does not create duplicates (it only tops up joins).
package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/community-platform/backend/internal/config"
	"github.com/community-platform/backend/internal/database"
	"github.com/community-platform/backend/internal/domain"
	"github.com/community-platform/backend/internal/repository"
	"github.com/jmoiron/sqlx"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		log.Fatalf("connect to postgres: %v — is `make docker-up` running?", err)
	}
	defer db.Close()

	ctx := context.Background()
	userRepo := repository.NewUserRepository(db)
	communityRepo := repository.NewCommunityRepository(db)
	campaignRepo := repository.NewCampaignRepository(db)

	fmt.Println("── Seeding CommunityDeals demo data ──────────────────────────")

	// ── 1. Users ────────────────────────────────────────────────────────────
	// The first user (Rahul) is the demo login for the web/mobile apps.
	// Login with this phone + OTP "111111" (OTP_DEV_MODE=true in .env).
	demoUsers := []struct {
		phone string
		name  string
		role  domain.UserRole
	}{
		{"+919810000001", "Rahul Sharma", domain.RoleCommunityAdmin},
		{"+919810000002", "Priya Deshmukh", domain.RoleMember},
		{"+919810000003", "Amit Kulkarni", domain.RoleMember},
		{"+919810000004", "Sana Iyer", domain.RoleMember},
		{"+919810000005", "Vikram Rao", domain.RoleMember},
		{"+919810000006", "Neha Joshi", domain.RoleMember},
		{"+919810000007", "Arjun Mehta", domain.RoleMember},
		{"+919810000008", "Kavya Nair", domain.RoleMember},
		{"+919810000009", "Rohan Verma", domain.RoleMember},
		{"+919810000010", "Ananya Gupta", domain.RoleMember},
		{"+919810000011", "Karan Malhotra", domain.RoleMember},
		{"+919810000012", "Divya Reddy", domain.RoleMember},
		{"+919810000013", "Siddharth Bose", domain.RoleMember},
		{"+919810000014", "Meera Pillai", domain.RoleMember},
		{"+919810000015", "Aditya Kapoor", domain.RoleMember},
		{"+919810000016", "Ishita Bhatt", domain.RoleMember},
		{"+919810000017", "Farhan Sheikh", domain.RoleMember},
		{"+919810000018", "Tanvi Rane", domain.RoleMember},
		{"+919810000019", "Manish Agarwal", domain.RoleMember},
		{"+919810000020", "Pooja Thakur", domain.RoleMember},
		{"+919810000021", "Gaurav Chopra", domain.RoleMember},
		{"+919810000022", "Ritu Saxena", domain.RoleMember},
		{"+919810000023", "Nikhil Bansal", domain.RoleMember},
		{"+919810000024", "Shreya Menon", domain.RoleMember},
		{"+919810000025", "Yash Trivedi", domain.RoleMember},
	}

	users := make([]*domain.User, 0, len(demoUsers))
	for _, du := range demoUsers {
		u, err := ensureUser(ctx, userRepo, du.phone, du.name, du.role)
		if err != nil {
			log.Fatalf("seed user %s: %v", du.name, err)
		}
		users = append(users, u)
	}
	fmt.Printf("✓ %d demo users ready\n", len(users))
	vendor := users[0] // Rahul — community admin / campaign creator

	// ── 2. Community ────────────────────────────────────────────────────────
	community, err := ensureCommunity(ctx, communityRepo, vendor.ID)
	if err != nil {
		log.Fatalf("seed community: %v", err)
	}
	fmt.Printf("✓ community %q ready\n", community.Name)

	for _, u := range users {
		if err := ensureMember(ctx, communityRepo, community.ID, u.ID); err != nil {
			log.Fatalf("add member %s: %v", u.Name, err)
		}
	}
	fmt.Printf("✓ %d members in %q\n", len(users), community.Name)

	// Extra societies in other cities that the demo user is NOT a member of,
	// so the "find your society" search has something to actually find.
	neighbours := []communitySeed{
		{name: "Sunrise Residency", inviteCode: "SNR2024", city: "Pune",
			state: "Maharashtra", pinCode: "411045", address: "Sunrise Residency, Wakad",
			description: "120-flat gated society in Wakad.", requiresApproval: false},
		{name: "Palm Grove Apartments", inviteCode: "PLM2024", city: "Pune",
			state: "Maharashtra", pinCode: "411057", address: "Palm Grove, Hinjewadi Phase 2",
			description: "80 flats next to the IT park.", requiresApproval: true},
		{name: "Lake View Heights", inviteCode: "LKV2024", city: "Mumbai",
			state: "Maharashtra", pinCode: "400076", address: "Lake View Heights, Powai",
			description: "Lakeside towers in Powai.", requiresApproval: false},
	}
	for _, n := range neighbours {
		c, err := ensureCommunityFrom(ctx, communityRepo, n, vendor.ID)
		if err != nil {
			log.Fatalf("seed community %s: %v", n.name, err)
		}
		// A few members each so they don't look abandoned in search results.
		for _, u := range users[20:] {
			if err := ensureMember(ctx, communityRepo, c.ID, u.ID); err != nil {
				log.Fatalf("add member to %s: %v", n.name, err)
			}
		}
	}
	fmt.Printf("✓ %d more societies to discover (Pune & Mumbai)\n", len(neighbours))

	// ── 3. Campaigns ────────────────────────────────────────────────────────
	acService, err := ensureCampaign(ctx, db, campaignRepo, campaignSeed{
		communityID: community.ID,
		vendorID:    vendor.ID,
		serviceName: "AC Service",
		title:       "AC Service This Sunday",
		description: "Annual AC servicing for all split & window units — " +
			"gas top-up extra if needed. Verified technician, doorstep service.",
		imageURL:        "https://images.unsplash.com/photo-1614963366795-973eb8748ebb?w=800",
		minParticipants: 20,
		maxParticipants: 0,
		serviceDaysOut:  5,
		endHoursOut:     72,
		tiers: []tierSeed{
			{1, 9, 99900},
			{10, 19, 69900},
			{20, 0, 39900},
		},
	})
	if err != nil {
		log.Fatalf("seed campaign AC Service: %v", err)
	}
	// 18/20 — deliberately just short of unlocking the best tier (matches
	// the "20 flats needed to unlock the group discount" reference screen).
	if err := joinUsers(ctx, campaignRepo, acService.ID, users[1:19]); err != nil {
		log.Fatalf("join AC Service: %v", err)
	}

	carWash, err := ensureCampaign(ctx, db, campaignRepo, campaignSeed{
		communityID: community.ID,
		vendorID:    vendor.ID,
		serviceName: "Car Wash",
		title:       "Weekend Car Wash Subscription",
		description: "Doorstep car wash every Saturday morning for a month. " +
			"Group booking unlocked — price is locked in for all four washes.",
		imageURL:        "https://images.unsplash.com/photo-1558618666-fcd25c85cd64?w=800",
		minParticipants: 10,
		maxParticipants: 60,
		serviceDaysOut:  2,
		endHoursOut:     24,
		tiers: []tierSeed{
			{1, 9, 99900},
			{10, 24, 59900},
			{25, 0, 39900},
		},
	})
	if err != nil {
		log.Fatalf("seed campaign Car Wash: %v", err)
	}
	// 15 joined — past the minimum, so this one shows the "unlocked" state.
	if err := joinUsers(ctx, campaignRepo, carWash.ID, users[2:17]); err != nil {
		log.Fatalf("join Car Wash: %v", err)
	}

	cleaning, err := ensureCampaign(ctx, db, campaignRepo, campaignSeed{
		communityID: community.ID,
		vendorID:    vendor.ID,
		serviceName: "Deep Cleaning",
		title:       "Diwali Deep Cleaning Drive",
		description: "Pre-Diwali deep cleaning for kitchens & balconies. " +
			"Early days — help this one reach its minimum!",
		imageURL:        "https://images.unsplash.com/photo-1581578731548-c64695cc6952?w=800",
		minParticipants: 15,
		maxParticipants: 0,
		serviceDaysOut:  10,
		endHoursOut:     168,
		tiers: []tierSeed{
			{1, 14, 149900},
			{15, 0, 89900},
		},
	})
	if err != nil {
		log.Fatalf("seed campaign Deep Cleaning: %v", err)
	}
	// Only 3 joined — a fresh, early-stage campaign.
	if err := joinUsers(ctx, campaignRepo, cleaning.ID, users[3:6]); err != nil {
		log.Fatalf("join Deep Cleaning: %v", err)
	}

	fmt.Println("✓ 3 demo campaigns ready (AC Service, Car Wash, Deep Cleaning)")
	fmt.Println()
	fmt.Println("── Demo login ──────────────────────────────────────────────")
	fmt.Println("Phone: 9810000001   (Rahul Sharma — community admin)")
	fmt.Println("OTP:   111111       (fixed in dev — OTP_DEV_MODE=true)")
	fmt.Println("Try also: 9810000002 … 9810000025 for a regular member view.")
	fmt.Println("────────────────────────────────────────────────────────────")
}

// ── helpers ──────────────────────────────────────────────────────────────────

type communitySeed struct {
	name, description, inviteCode string
	city, state, pinCode, address string
	requiresApproval              bool
}

type tierSeed struct {
	min, max int
	price    int64 // paise
}

type campaignSeed struct {
	communityID, vendorID            string
	serviceName, title, description  string
	imageURL                         string
	minParticipants, maxParticipants int
	serviceDaysOut, endHoursOut      int
	tiers                            []tierSeed
}

func ensureUser(ctx context.Context, repo repository.UserRepository, phone, name string, role domain.UserRole) (*domain.User, error) {
	if u, err := repo.FindByPhone(ctx, phone); err == nil {
		return u, nil
	}
	now := time.Now()
	u := &domain.User{
		ID:         newID(),
		Phone:      phone,
		Name:       name,
		AvatarURL:  fmt.Sprintf("https://i.pravatar.cc/150?u=%s", phone),
		Role:       role,
		Status:     domain.UserStatusActive,
		IsVerified: true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func ensureCommunity(ctx context.Context, repo repository.CommunityRepository, createdByID string) (*domain.Community, error) {
	return ensureCommunityFrom(ctx, repo, communitySeed{
		name:        "Green Valley Society",
		description: "A 200-flat residential society in Baner, Pune.",
		inviteCode:  "GVS2024",
		city:        "Pune",
		state:       "Maharashtra",
		pinCode:     "411045",
		address:     "Green Valley Society, Baner Road",
	}, createdByID)
}

func ensureCommunityFrom(ctx context.Context, repo repository.CommunityRepository, s communitySeed, createdByID string) (*domain.Community, error) {
	if c, err := repo.FindByInviteCode(ctx, s.inviteCode); err == nil {
		return c, nil
	}
	now := time.Now()
	c := &domain.Community{
		ID:               newID(),
		Name:             s.name,
		Description:      s.description,
		Type:             domain.CommunityTypeSociety,
		Status:           domain.CommunityStatusActive,
		City:             s.city,
		State:            s.state,
		PinCode:          s.pinCode,
		Address:          s.address,
		RequiresApproval: s.requiresApproval,
		InviteCode:       s.inviteCode,
		CreatedByID:      createdByID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func ensureMember(ctx context.Context, repo repository.CommunityRepository, communityID, userID string) error {
	if _, err := repo.FindMember(ctx, communityID, userID); err == nil {
		return nil
	}
	now := time.Now()
	m := &domain.CommunityMember{
		ID:          newID(),
		CommunityID: communityID,
		UserID:      userID,
		Role:        domain.MemberRoleMember,
		Status:      domain.MemberStatusApproved,
		JoinedAt:    &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err := repo.AddMember(ctx, m)
	if err != nil && err != repository.ErrAlreadyMember {
		return err
	}
	return nil
}

func ensureCampaign(ctx context.Context, db *sqlx.DB, repo repository.CampaignRepository, s campaignSeed) (*domain.Campaign, error) {
	existing, _, err := repo.List(ctx, repository.CampaignFilter{CommunityID: s.communityID, Page: 1, Limit: 50})
	if err == nil {
		for _, c := range existing {
			if c.Title == s.title {
				// Re-running the seed days later would otherwise leave the demo
				// with lapsed deadlines, and an expired campaign can't be joined.
				// Slide the dates forward so the demo is always live.
				if err := refreshCampaignDates(ctx, db, c, s); err != nil {
					return nil, fmt.Errorf("refreshing dates: %w", err)
				}
				return c, nil
			}
		}
	}

	now := time.Now()
	c := &domain.Campaign{
		ID:               newID(),
		CommunityID:      s.communityID,
		VendorID:         s.vendorID,
		ServiceName:      s.serviceName,
		Title:            s.title,
		Description:      s.description,
		ImageURL:         s.imageURL,
		MinParticipants:  s.minParticipants,
		MaxParticipants:  s.maxParticipants,
		ServiceDate:      now.Add(time.Duration(s.serviceDaysOut) * 24 * time.Hour),
		StartDate:        now.Add(-24 * time.Hour),
		EndDate:          now.Add(time.Duration(s.endHoursOut) * time.Hour),
		Status:           domain.CampaignStatusPublished,
		ParticipantCount: 0,
		Slug:             slugify(s.title) + "-" + shortID(),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := repo.Create(ctx, c); err != nil {
		return nil, err
	}

	tiers := make([]*domain.CampaignPricingTier, len(s.tiers))
	for i, t := range s.tiers {
		tiers[i] = &domain.CampaignPricingTier{
			ID:         newID(),
			CampaignID: c.ID,
			MinCount:   t.min,
			MaxCount:   t.max,
			Price:      t.price,
			TierOrder:  i + 1,
		}
	}
	if err := repo.CreatePricingTiers(ctx, tiers); err != nil {
		return nil, err
	}
	return c, nil
}

// refreshCampaignDates slides an existing demo campaign's window to be
// relative to now, and clears EXPIRED so it becomes joinable again.
func refreshCampaignDates(ctx context.Context, db *sqlx.DB, c *domain.Campaign, s campaignSeed) error {
	now := time.Now()
	_, err := db.ExecContext(ctx, `
		UPDATE campaigns
		SET service_date = $1,
		    start_date   = $2,
		    end_date     = $3,
		    status       = CASE WHEN status = 'EXPIRED' THEN 'PUBLISHED' ELSE status END,
		    updated_at   = $4
		WHERE id = $5
	`,
		now.Add(time.Duration(s.serviceDaysOut)*24*time.Hour),
		now.Add(-24*time.Hour),
		now.Add(time.Duration(s.endHoursOut)*time.Hour),
		now,
		c.ID,
	)
	if err != nil {
		return err
	}
	c.EndDate = now.Add(time.Duration(s.endHoursOut) * time.Hour)
	return nil
}

func joinUsers(ctx context.Context, repo repository.CampaignRepository, campaignID string, users []*domain.User) error {
	for _, u := range users {
		_, _, err := repo.JoinCampaign(ctx, campaignID, u.ID)
		if err != nil && err != repository.ErrAlreadyJoined {
			return fmt.Errorf("user %s: %w", u.Name, err)
		}
	}
	return nil
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func shortID() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%06x", b)[:6]
}

func slugify(title string) string {
	out := make([]rune, 0, len(title))
	lastDash := false
	for _, r := range title {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			out = append(out, r)
			lastDash = false
		case r >= 'A' && r <= 'Z':
			out = append(out, r+32)
			lastDash = false
		default:
			if !lastDash && len(out) > 0 {
				out = append(out, '-')
				lastDash = true
			}
		}
	}
	s := string(out)
	if len(s) > 0 && s[len(s)-1] == '-' {
		s = s[:len(s)-1]
	}
	return s
}
