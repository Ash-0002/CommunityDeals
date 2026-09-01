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

// ErrNotFound is returned when a record does not exist.
var ErrNotFound = errors.New("record not found")

// ErrDuplicatePhone is returned when a phone number is already registered.
var ErrDuplicatePhone = errors.New("phone number already registered")

// UserRepository defines data access operations for User.
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByPhone(ctx context.Context, phone string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	SetLastLogin(ctx context.Context, userID string, at time.Time) error

	// Refresh tokens
	SaveRefreshToken(ctx context.Context, rt *domain.RefreshToken) error
	FindRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllUserTokens(ctx context.Context, userID string) error
}

type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a PostgreSQL-backed UserRepository.
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	query := `SELECT * FROM users WHERE id = $1 AND status != 'BANNED' LIMIT 1`
	if err := r.db.GetContext(ctx, &user, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &user, nil
}

func (r *userRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	var user domain.User
	query := `SELECT * FROM users WHERE phone = $1 LIMIT 1`
	if err := r.db.GetContext(ctx, &user, query, phone); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find user by phone: %w", err)
	}
	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, phone, name, email, avatar_url, role, status, is_verified, created_at, updated_at)
		VALUES (:id, :phone, :name, :email, :avatar_url, :role, :status, :is_verified, :created_at, :updated_at)
	`
	if _, err := r.db.NamedExecContext(ctx, query, user); err != nil {
		// Check unique violation on phone
		if isUniqueViolation(err) {
			return ErrDuplicatePhone
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = time.Now()
	query := `
		UPDATE users
		SET name = :name, email = :email, avatar_url = :avatar_url, updated_at = :updated_at
		WHERE id = :id
	`
	if _, err := r.db.NamedExecContext(ctx, query, user); err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (r *userRepository) SetLastLogin(ctx context.Context, userID string, at time.Time) error {
	query := `UPDATE users SET last_login_at = $1, updated_at = $2 WHERE id = $3`
	if _, err := r.db.ExecContext(ctx, query, at, at, userID); err != nil {
		return fmt.Errorf("set last login: %w", err)
	}
	return nil
}

func (r *userRepository) SaveRefreshToken(ctx context.Context, rt *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
		VALUES (:id, :user_id, :token, :expires_at, :created_at)
	`
	if _, err := r.db.NamedExecContext(ctx, query, rt); err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}
	return nil
}

func (r *userRepository) FindRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var rt domain.RefreshToken
	query := `SELECT * FROM refresh_tokens WHERE token = $1 LIMIT 1`
	if err := r.db.GetContext(ctx, &rt, query, tokenHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find refresh token: %w", err)
	}
	return &rt, nil
}

func (r *userRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	now := time.Now()
	query := `UPDATE refresh_tokens SET revoked_at = $1 WHERE token = $2`
	if _, err := r.db.ExecContext(ctx, query, now, tokenHash); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

func (r *userRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	now := time.Now()
	query := `UPDATE refresh_tokens SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`
	if _, err := r.db.ExecContext(ctx, query, now, userID); err != nil {
		return fmt.Errorf("revoke all user tokens: %w", err)
	}
	return nil
}

// isUniqueViolation checks if a pq error is a unique constraint violation (code 23505).
func isUniqueViolation(err error) bool {
	return err != nil && (err.Error() != "" && len(err.Error()) > 0 &&
		// lib/pq error code for unique_violation
		fmt.Sprintf("%v", err) != "" &&
		containsString(err.Error(), "23505"))
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > 0 && searchStr(s, substr))
}

func searchStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
