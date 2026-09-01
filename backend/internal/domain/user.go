package domain

import (
	"time"
)

// UserRole defines what a user can do on the platform.
type UserRole string

const (
	RoleMember        UserRole = "MEMBER"
	RoleCommunityAdmin UserRole = "COMMUNITY_ADMIN"
	RoleVendor        UserRole = "VENDOR"
	RolePlatformAdmin UserRole = "PLATFORM_ADMIN"
)

// UserStatus tracks whether the account is usable.
type UserStatus string

const (
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusInactive UserStatus = "INACTIVE"
	UserStatusBanned   UserStatus = "BANNED"
)

// User is the central identity entity. Authentication is phone-based OTP.
type User struct {
	ID          string     `db:"id"`
	Phone       string     `db:"phone"`       // E.164 format, e.g. +919876543210
	Name        string     `db:"name"`
	Email       string     `db:"email"`       // optional
	AvatarURL   string     `db:"avatar_url"`  // optional
	Role        UserRole   `db:"role"`
	Status      UserStatus `db:"status"`
	IsVerified  bool       `db:"is_verified"` // true once OTP verified at least once
	LastLoginAt *time.Time `db:"last_login_at"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

// IsActive checks if the user can use the platform.
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsPlatformAdmin checks if the user has platform-wide admin privileges.
func (u *User) IsPlatformAdmin() bool {
	return u.Role == RolePlatformAdmin
}

// RefreshToken stores a refresh token associated with a user session.
type RefreshToken struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Token     string    `db:"token"`      // hashed
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
	RevokedAt *time.Time `db:"revoked_at"` // nil = still valid
}

// IsValid returns true when the token has not expired and not been revoked.
func (rt *RefreshToken) IsValid() bool {
	return rt.RevokedAt == nil && time.Now().Before(rt.ExpiresAt)
}
