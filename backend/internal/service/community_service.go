package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/community-platform/backend/internal/domain"
	"github.com/community-platform/backend/internal/dto"
	"github.com/community-platform/backend/internal/repository"
)

// Sentinel errors for community flows.
var (
	ErrCommunityNotFound   = errors.New("community not found")
	ErrAlreadyMember       = errors.New("you are already a member of this community")
	ErrNotCommunityMember  = errors.New("you are not a member of this community")
	ErrNotCommunityAdmin   = errors.New("you do not have admin rights in this community")
	ErrInvalidInviteCode   = errors.New("invalid invite code")
	ErrCommunityInactive   = errors.New("community is not active")
)

// CommunityService handles community business logic.
type CommunityService interface {
	Create(ctx context.Context, creatorID string, req dto.CreateCommunityRequest) (*dto.CommunityResponse, error)
	GetByID(ctx context.Context, id string, requestingUserID string) (*dto.CommunityResponse, error)
	List(ctx context.Context, filter dto.ListCommunitiesQuery) (*dto.PaginatedCommunitiesResponse, error)
	Join(ctx context.Context, communityID, userID string, req dto.JoinCommunityRequest) (*dto.JoinCommunityResponse, error)
	Leave(ctx context.Context, communityID, userID string) error
	ListMembers(ctx context.Context, communityID, requestingUserID string, page, limit int) ([]dto.CommunityMemberResponse, int, error)
	ApproveMember(ctx context.Context, communityID, adminID, targetUserID string) error
	ListUserCommunities(ctx context.Context, userID string) ([]dto.CommunityResponse, error)
}

type communityService struct {
	communityRepo repository.CommunityRepository
	userRepo      repository.UserRepository
}

// NewCommunityService creates a CommunityService with all dependencies injected.
func NewCommunityService(communityRepo repository.CommunityRepository, userRepo repository.UserRepository) CommunityService {
	return &communityService{
		communityRepo: communityRepo,
		userRepo:      userRepo,
	}
}

// Create creates a new community and automatically makes the creator an admin.
func (s *communityService) Create(ctx context.Context, creatorID string, req dto.CreateCommunityRequest) (*dto.CommunityResponse, error) {
	// Validate community type
	if !isValidCommunityType(req.Type) {
		return nil, fmt.Errorf("invalid community type: %s", req.Type)
	}

	inviteCode, err := generateInviteCode()
	if err != nil {
		return nil, fmt.Errorf("generating invite code: %w", err)
	}

	now := time.Now()
	community := &domain.Community{
		ID:               newUUID(),
		Name:             strings.TrimSpace(req.Name),
		Description:      strings.TrimSpace(req.Description),
		Type:             domain.CommunityType(req.Type),
		Status:           domain.CommunityStatusActive,
		City:             strings.TrimSpace(req.City),
		State:            strings.TrimSpace(req.State),
		PinCode:          strings.TrimSpace(req.PinCode),
		Address:          strings.TrimSpace(req.Address),
		RequiresApproval: req.RequiresApproval,
		InviteCode:       inviteCode,
		CreatedByID:      creatorID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.communityRepo.Create(ctx, community); err != nil {
		return nil, fmt.Errorf("creating community: %w", err)
	}

	// Automatically make creator an approved admin
	joinedAt := now
	adminMember := &domain.CommunityMember{
		ID:          newUUID(),
		CommunityID: community.ID,
		UserID:      creatorID,
		Role:        domain.MemberRoleAdmin,
		Status:      domain.MemberStatusApproved,
		JoinedAt:    &joinedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.communityRepo.AddMember(ctx, adminMember); err != nil {
		return nil, fmt.Errorf("adding creator as admin: %w", err)
	}

	resp := toCommunityResponse(community, 1, true) // creator is admin, so show invite code
	return &resp, nil
}

// GetByID returns community details. Invite code is only shown to admins.
func (s *communityService) GetByID(ctx context.Context, id, requestingUserID string) (*dto.CommunityResponse, error) {
	community, err := s.communityRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCommunityNotFound
		}
		return nil, fmt.Errorf("getting community: %w", err)
	}

	memberCount, _ := s.communityRepo.CountMembers(ctx, id)

	isAdmin := false
	if requestingUserID != "" {
		isAdmin, _ = s.communityRepo.IsAdmin(ctx, id, requestingUserID)
	}

	resp := toCommunityResponse(community, memberCount, isAdmin)
	return &resp, nil
}

// List returns a paginated, filtered list of active communities.
func (s *communityService) List(ctx context.Context, q dto.ListCommunitiesQuery) (*dto.PaginatedCommunitiesResponse, error) {
	filter := repository.CommunityFilter{
		Type:    q.Type,
		City:    q.City,
		PinCode: q.PinCode,
		Page:    q.Page,
		Limit:   q.Limit,
	}

	communities, total, err := s.communityRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing communities: %w", err)
	}

	responses := make([]dto.CommunityResponse, 0, len(communities))
	for _, c := range communities {
		count, _ := s.communityRepo.CountMembers(ctx, c.ID)
		responses = append(responses, toCommunityResponse(c, count, false))
	}

	return &dto.PaginatedCommunitiesResponse{
		Communities: responses,
		Total:       total,
		Page:        q.Page,
		Limit:       q.Limit,
		HasMore:     (q.Page * q.Limit) < total,
	}, nil
}

// Join processes a user's request to join a community.
// If the community requires approval, the status will be PENDING.
// If the community uses an invite code, that code must be provided and match.
func (s *communityService) Join(ctx context.Context, communityID, userID string, req dto.JoinCommunityRequest) (*dto.JoinCommunityResponse, error) {
	community, err := s.communityRepo.FindByID(ctx, communityID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCommunityNotFound
		}
		return nil, fmt.Errorf("finding community: %w", err)
	}

	if !community.IsActive() {
		return nil, ErrCommunityInactive
	}

	// Validate invite code if one is set on the community
	if community.InviteCode != "" && req.InviteCode != "" {
		if req.InviteCode != community.InviteCode {
			return nil, ErrInvalidInviteCode
		}
	}

	// Check for existing membership (any status)
	existing, err := s.communityRepo.FindMember(ctx, communityID, userID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("checking membership: %w", err)
	}
	if existing != nil {
		if existing.Status == domain.MemberStatusApproved {
			return nil, ErrAlreadyMember
		}
		if existing.Status == domain.MemberStatusPending {
			return &dto.JoinCommunityResponse{
				MembershipStatus: "PENDING",
				Message:          "Your join request is already pending admin approval.",
			}, nil
		}
		// REMOVED or REJECTED — allow re-join by updating status
		status := domain.MemberStatusApproved
		msg := "You have rejoined the community."
		if community.RequiresApproval {
			status = domain.MemberStatusPending
			msg = "Your join request is pending admin approval."
		}
		if err := s.communityRepo.UpdateMemberStatus(ctx, communityID, userID, status); err != nil {
			return nil, fmt.Errorf("re-joining community: %w", err)
		}
		return &dto.JoinCommunityResponse{MembershipStatus: string(status), Message: msg}, nil
	}

	// New membership
	now := time.Now()
	status := domain.MemberStatusApproved
	var joinedAt *time.Time
	if !community.RequiresApproval {
		joinedAt = &now
	} else {
		status = domain.MemberStatusPending
	}

	member := &domain.CommunityMember{
		ID:          newUUID(),
		CommunityID: communityID,
		UserID:      userID,
		Role:        domain.MemberRoleMember,
		Status:      status,
		JoinedAt:    joinedAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.communityRepo.AddMember(ctx, member); err != nil {
		if errors.Is(err, repository.ErrAlreadyMember) {
			return nil, ErrAlreadyMember
		}
		return nil, fmt.Errorf("adding member: %w", err)
	}

	msg := "You have successfully joined the community."
	if status == domain.MemberStatusPending {
		msg = "Your join request is pending admin approval."
	}

	return &dto.JoinCommunityResponse{
		MembershipStatus: string(status),
		Message:          msg,
	}, nil
}

// Leave removes the user from a community.
func (s *communityService) Leave(ctx context.Context, communityID, userID string) error {
	member, err := s.communityRepo.FindMember(ctx, communityID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotCommunityMember
		}
		return fmt.Errorf("finding member: %w", err)
	}

	if !member.IsActive() {
		return ErrNotCommunityMember
	}

	// Prevent the last admin from leaving
	if member.IsAdmin() {
		isOnlyAdmin, err := s.isOnlyAdmin(ctx, communityID, userID)
		if err != nil {
			return err
		}
		if isOnlyAdmin {
			return fmt.Errorf("you are the only admin — assign another admin before leaving")
		}
	}

	return s.communityRepo.RemoveMember(ctx, communityID, userID)
}

// ListMembers returns approved members of a community. Requires membership.
func (s *communityService) ListMembers(ctx context.Context, communityID, requestingUserID string, page, limit int) ([]dto.CommunityMemberResponse, int, error) {
	isMember, err := s.communityRepo.IsMember(ctx, communityID, requestingUserID)
	if err != nil {
		return nil, 0, err
	}
	if !isMember {
		return nil, 0, ErrNotCommunityMember
	}

	members, total, err := s.communityRepo.ListMembers(ctx, communityID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("listing members: %w", err)
	}

	responses := make([]dto.CommunityMemberResponse, 0, len(members))
	for _, m := range members {
		// Fetch user details for each member
		user, err := s.userRepo.FindByID(ctx, m.UserID)
		if err != nil {
			continue // skip if user somehow deleted
		}
		joinedAt := ""
		if m.JoinedAt != nil {
			joinedAt = m.JoinedAt.Format(time.RFC3339)
		}
		responses = append(responses, dto.CommunityMemberResponse{
			UserID:   m.UserID,
			Name:     user.Name,
			Phone:    user.Phone,
			Role:     string(m.Role),
			Status:   string(m.Status),
			JoinedAt: joinedAt,
		})
	}

	return responses, total, nil
}

// ApproveMember approves a pending join request. Only community admins can do this.
func (s *communityService) ApproveMember(ctx context.Context, communityID, adminID, targetUserID string) error {
	isAdmin, err := s.communityRepo.IsAdmin(ctx, communityID, adminID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return ErrNotCommunityAdmin
	}

	return s.communityRepo.UpdateMemberStatus(ctx, communityID, targetUserID, domain.MemberStatusApproved)
}

// ListUserCommunities returns all communities the user is an approved member of.
func (s *communityService) ListUserCommunities(ctx context.Context, userID string) ([]dto.CommunityResponse, error) {
	communities, err := s.communityRepo.ListUserCommunities(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing user communities: %w", err)
	}

	responses := make([]dto.CommunityResponse, 0, len(communities))
	for _, c := range communities {
		count, _ := s.communityRepo.CountMembers(ctx, c.ID)
		isAdmin, _ := s.communityRepo.IsAdmin(ctx, c.ID, userID)
		responses = append(responses, toCommunityResponse(c, count, isAdmin))
	}
	return responses, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (s *communityService) isOnlyAdmin(ctx context.Context, communityID, userID string) (bool, error) {
	members, _, err := s.communityRepo.ListMembers(ctx, communityID, 1, 100)
	if err != nil {
		return false, err
	}
	adminCount := 0
	for _, m := range members {
		if m.Role == domain.MemberRoleAdmin && m.Status == domain.MemberStatusApproved {
			adminCount++
		}
	}
	return adminCount <= 1, nil
}

func toCommunityResponse(c *domain.Community, memberCount int, showInviteCode bool) dto.CommunityResponse {
	inviteCode := ""
	if showInviteCode {
		inviteCode = c.InviteCode
	}
	return dto.CommunityResponse{
		ID:               c.ID,
		Name:             c.Name,
		Description:      c.Description,
		Type:             string(c.Type),
		Status:           string(c.Status),
		City:             c.City,
		State:            c.State,
		PinCode:          c.PinCode,
		Address:          c.Address,
		RequiresApproval: c.RequiresApproval,
		InviteCode:       inviteCode,
		LogoURL:          c.LogoURL,
		MemberCount:      memberCount,
		CreatedByID:      c.CreatedByID,
		CreatedAt:        c.CreatedAt.Format(time.RFC3339),
	}
}

// generateInviteCode returns a short alphanumeric code like "GV4K2X".
func generateInviteCode() (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no confusable chars
	const length = 6
	code := make([]byte, length)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[n.Int64()]
	}
	return string(code), nil
}

func isValidCommunityType(t string) bool {
	valid := map[string]bool{
		"SOCIETY": true, "APARTMENT": true, "RESIDENTIAL_COMPLEX": true,
		"OFFICE": true, "COLLEGE": true, "CORPORATE_GROUP": true,
		"PRIVATE_GROUP": true, "PARTNER_COMMUNITY": true,
	}
	return valid[t]
}
