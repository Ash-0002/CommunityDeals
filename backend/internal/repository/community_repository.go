package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/community-platform/backend/internal/domain"
	"github.com/jmoiron/sqlx"
)

// ErrAlreadyMember is returned when a user tries to join a community they already belong to.
var ErrAlreadyMember = errors.New("user is already a member of this community")

// ErrNotMember is returned when an action requires membership.
var ErrNotMember = errors.New("user is not a member of this community")

// ErrInvalidInviteCode is returned when an invite code doesn't match.
var ErrInvalidInviteCode = errors.New("invalid invite code")

// CommunityRepository defines data-access operations for Community.
type CommunityRepository interface {
	Create(ctx context.Context, c *domain.Community) error
	FindByID(ctx context.Context, id string) (*domain.Community, error)
	FindByInviteCode(ctx context.Context, code string) (*domain.Community, error)
	List(ctx context.Context, filter CommunityFilter) ([]*domain.Community, int, error)
	Update(ctx context.Context, c *domain.Community) error

	// Membership
	AddMember(ctx context.Context, m *domain.CommunityMember) error
	FindMember(ctx context.Context, communityID, userID string) (*domain.CommunityMember, error)
	UpdateMemberStatus(ctx context.Context, communityID, userID string, status domain.MemberStatus) error
	RemoveMember(ctx context.Context, communityID, userID string) error
	ListMembers(ctx context.Context, communityID string, page, limit int) ([]*domain.CommunityMember, int, error)
	CountMembers(ctx context.Context, communityID string) (int, error)
	IsMember(ctx context.Context, communityID, userID string) (bool, error)
	IsAdmin(ctx context.Context, communityID, userID string) (bool, error)

	// User's communities
	ListUserCommunities(ctx context.Context, userID string) ([]*domain.Community, error)
}

// CommunityFilter is used to filter the community list query.
type CommunityFilter struct {
	Type    string
	City    string
	PinCode string
	Page    int
	Limit   int
}

type communityRepository struct {
	db *sqlx.DB
}

// NewCommunityRepository creates a PostgreSQL-backed CommunityRepository.
func NewCommunityRepository(db *sqlx.DB) CommunityRepository {
	return &communityRepository{db: db}
}

// ── Community CRUD ─────────────────────────────────────────────────────────────

func (r *communityRepository) Create(ctx context.Context, c *domain.Community) error {
	query := `
		INSERT INTO communities
			(id, name, description, type, status, city, state, pin_code, address,
			 requires_approval, invite_code, logo_url, created_by_id, created_at, updated_at)
		VALUES
			(:id, :name, :description, :type, :status, :city, :state, :pin_code, :address,
			 :requires_approval, :invite_code, :logo_url, :created_by_id, :created_at, :updated_at)
	`
	if _, err := r.db.NamedExecContext(ctx, query, c); err != nil {
		return fmt.Errorf("create community: %w", err)
	}
	return nil
}

func (r *communityRepository) FindByID(ctx context.Context, id string) (*domain.Community, error) {
	var c domain.Community
	query := `SELECT * FROM communities WHERE id = $1 LIMIT 1`
	if err := r.db.GetContext(ctx, &c, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find community by id: %w", err)
	}
	return &c, nil
}

func (r *communityRepository) FindByInviteCode(ctx context.Context, code string) (*domain.Community, error) {
	var c domain.Community
	query := `SELECT * FROM communities WHERE invite_code = $1 AND status = 'ACTIVE' LIMIT 1`
	if err := r.db.GetContext(ctx, &c, query, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find community by invite code: %w", err)
	}
	return &c, nil
}

func (r *communityRepository) List(ctx context.Context, f CommunityFilter) ([]*domain.Community, int, error) {
	// Build dynamic WHERE clause
	where := "WHERE status = 'ACTIVE'"
	args := []interface{}{}
	argIdx := 1

	if f.Type != "" {
		where += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, f.Type)
		argIdx++
	}
	if f.City != "" {
		where += fmt.Sprintf(" AND LOWER(city) = LOWER($%d)", argIdx)
		args = append(args, f.City)
		argIdx++
	}
	if f.PinCode != "" {
		where += fmt.Sprintf(" AND pin_code = $%d", argIdx)
		args = append(args, f.PinCode)
		argIdx++
	}

	// Count total
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM communities %s", where)
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("count communities: %w", err)
	}

	// Paginate
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	offset := (f.Page - 1) * f.Limit

	listQuery := fmt.Sprintf(
		"SELECT * FROM communities %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		where, argIdx, argIdx+1,
	)
	args = append(args, f.Limit, offset)

	var communities []*domain.Community
	if err := r.db.SelectContext(ctx, &communities, listQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("list communities: %w", err)
	}

	return communities, total, nil
}

func (r *communityRepository) Update(ctx context.Context, c *domain.Community) error {
	c.UpdatedAt = time.Now()
	query := `
		UPDATE communities
		SET name = :name, description = :description, city = :city,
		    state = :state, pin_code = :pin_code, address = :address,
		    requires_approval = :requires_approval, logo_url = :logo_url,
		    status = :status, updated_at = :updated_at
		WHERE id = :id
	`
	if _, err := r.db.NamedExecContext(ctx, query, c); err != nil {
		return fmt.Errorf("update community: %w", err)
	}
	return nil
}

// ── Membership ─────────────────────────────────────────────────────────────────

func (r *communityRepository) AddMember(ctx context.Context, m *domain.CommunityMember) error {
	query := `
		INSERT INTO community_members
			(id, community_id, user_id, role, status, joined_at, created_at, updated_at)
		VALUES
			(:id, :community_id, :user_id, :role, :status, :joined_at, :created_at, :updated_at)
	`
	if _, err := r.db.NamedExecContext(ctx, query, m); err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyMember
		}
		return fmt.Errorf("add community member: %w", err)
	}
	return nil
}

func (r *communityRepository) FindMember(ctx context.Context, communityID, userID string) (*domain.CommunityMember, error) {
	var m domain.CommunityMember
	query := `SELECT * FROM community_members WHERE community_id = $1 AND user_id = $2 LIMIT 1`
	if err := r.db.GetContext(ctx, &m, query, communityID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find community member: %w", err)
	}
	return &m, nil
}

func (r *communityRepository) UpdateMemberStatus(ctx context.Context, communityID, userID string, status domain.MemberStatus) error {
	now := time.Now()
	var joinedAt *time.Time
	if status == domain.MemberStatusApproved {
		joinedAt = &now
	}
	query := `
		UPDATE community_members
		SET status = $1, joined_at = COALESCE($2, joined_at), updated_at = $3
		WHERE community_id = $4 AND user_id = $5
	`
	if _, err := r.db.ExecContext(ctx, query, status, joinedAt, now, communityID, userID); err != nil {
		return fmt.Errorf("update member status: %w", err)
	}
	return nil
}

func (r *communityRepository) RemoveMember(ctx context.Context, communityID, userID string) error {
	return r.UpdateMemberStatus(ctx, communityID, userID, domain.MemberStatusRemoved)
}

func (r *communityRepository) ListMembers(ctx context.Context, communityID string, page, limit int) ([]*domain.CommunityMember, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM community_members WHERE community_id = $1 AND status = 'APPROVED'`,
		communityID,
	); err != nil {
		return nil, 0, fmt.Errorf("count members: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var members []*domain.CommunityMember
	query := `
		SELECT * FROM community_members
		WHERE community_id = $1 AND status = 'APPROVED'
		ORDER BY joined_at ASC
		LIMIT $2 OFFSET $3
	`
	if err := r.db.SelectContext(ctx, &members, query, communityID, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("list members: %w", err)
	}
	return members, total, nil
}

func (r *communityRepository) CountMembers(ctx context.Context, communityID string) (int, error) {
	var count int
	if err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM community_members WHERE community_id = $1 AND status = 'APPROVED'`,
		communityID,
	); err != nil {
		return 0, fmt.Errorf("count members: %w", err)
	}
	return count, nil
}

func (r *communityRepository) IsMember(ctx context.Context, communityID, userID string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM community_members WHERE community_id=$1 AND user_id=$2 AND status='APPROVED'`,
		communityID, userID,
	)
	return count > 0, err
}

func (r *communityRepository) IsAdmin(ctx context.Context, communityID, userID string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM community_members WHERE community_id=$1 AND user_id=$2 AND role='ADMIN' AND status='APPROVED'`,
		communityID, userID,
	)
	return count > 0, err
}

func (r *communityRepository) ListUserCommunities(ctx context.Context, userID string) ([]*domain.Community, error) {
	var communities []*domain.Community
	query := `
		SELECT c.* FROM communities c
		INNER JOIN community_members cm ON c.id = cm.community_id
		WHERE cm.user_id = $1 AND cm.status = 'APPROVED' AND c.status = 'ACTIVE'
		ORDER BY cm.joined_at DESC
	`
	if err := r.db.SelectContext(ctx, &communities, query, userID); err != nil {
		return nil, fmt.Errorf("list user communities: %w", err)
	}
	return communities, nil
}
