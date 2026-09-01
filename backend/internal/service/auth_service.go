package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/community-platform/backend/internal/config"
	"github.com/community-platform/backend/internal/domain"
	"github.com/community-platform/backend/internal/dto"
	"github.com/community-platform/backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// Sentinel errors for auth flows.
var (
	ErrInvalidOTP      = errors.New("invalid or expired OTP")
	ErrRateLimited     = errors.New("too many OTP requests, please wait")
	ErrUserBanned      = errors.New("account is banned")
	ErrInvalidToken    = errors.New("invalid or expired token")
)

// otpRedisKey returns the Redis key for a phone's OTP.
func otpRedisKey(phone string) string {
	return fmt.Sprintf("otp:%s", phone)
}

// otpAttemptsKey returns the Redis key tracking failed OTP attempts.
func otpAttemptsKey(phone string) string {
	return fmt.Sprintf("otp_attempts:%s", phone)
}

// AuthService handles OTP-based authentication and JWT lifecycle.
type AuthService interface {
	SendOTP(ctx context.Context, req dto.SendOTPRequest) (*dto.SendOTPResponse, error)
	VerifyOTP(ctx context.Context, req dto.VerifyOTPRequest) (*dto.AuthTokensResponse, error)
	RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (*dto.AuthTokensResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	GetProfile(ctx context.Context, userID string) (*dto.UserResponse, error)
	UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequest) (*dto.UserResponse, error)
}

type authService struct {
	userRepo repository.UserRepository
	redis    *redis.Client
	cfg      *config.Config
}

// NewAuthService creates an AuthService with all dependencies injected.
func NewAuthService(userRepo repository.UserRepository, rdb *redis.Client, cfg *config.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		redis:    rdb,
		cfg:      cfg,
	}
}

// --- OTP ---

func (s *authService) SendOTP(ctx context.Context, req dto.SendOTPRequest) (*dto.SendOTPResponse, error) {
	phone := normalizePhone(req.Phone)

	// Rate limiting: allow max 5 OTPs per phone per hour
	attemptsKey := otpAttemptsKey(phone)
	count, _ := s.redis.Get(ctx, attemptsKey).Int()
	if count >= 5 {
		return nil, ErrRateLimited
	}

	// Generate OTP (6 digits)
	otp, err := generateOTP(6)
	if err != nil {
		return nil, fmt.Errorf("generating otp: %w", err)
	}

	// In dev mode always use "111111" for easy testing
	if s.cfg.OTP.DevMode {
		otp = "111111"
	}

	// Store OTP in Redis with expiry
	expiry := time.Duration(s.cfg.OTP.ExpiryMinutes) * time.Minute
	if err := s.redis.Set(ctx, otpRedisKey(phone), otp, expiry).Err(); err != nil {
		return nil, fmt.Errorf("storing otp: %w", err)
	}

	// Increment attempt counter
	pipe := s.redis.Pipeline()
	pipe.Incr(ctx, attemptsKey)
	pipe.Expire(ctx, attemptsKey, time.Hour)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("updating rate limit: %w", err)
	}

	// TODO: send SMS via Twilio/MSG91 when not in dev mode

	return &dto.SendOTPResponse{
		Phone:            phone,
		Message:          "OTP sent successfully",
		ExpiresInSeconds: s.cfg.OTP.ExpiryMinutes * 60,
	}, nil
}

func (s *authService) VerifyOTP(ctx context.Context, req dto.VerifyOTPRequest) (*dto.AuthTokensResponse, error) {
	phone := normalizePhone(req.Phone)

	// Retrieve stored OTP from Redis
	stored, err := s.redis.Get(ctx, otpRedisKey(phone)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrInvalidOTP
		}
		return nil, fmt.Errorf("getting otp from redis: %w", err)
	}

	if stored != req.OTP {
		return nil, ErrInvalidOTP
	}

	// OTP verified — delete it so it can't be reused
	_ = s.redis.Del(ctx, otpRedisKey(phone))
	_ = s.redis.Del(ctx, otpAttemptsKey(phone))

	// Find existing user or create new one
	user, err := s.userRepo.FindByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// First time — create account
			user = &domain.User{
				ID:         newUUID(),
				Phone:      phone,
				Role:       domain.RoleMember,
				Status:     domain.UserStatusActive,
				IsVerified: true,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}
			if err := s.userRepo.Create(ctx, user); err != nil {
				return nil, fmt.Errorf("creating user: %w", err)
			}
		} else {
			return nil, fmt.Errorf("finding user: %w", err)
		}
	}

	if !user.IsActive() {
		return nil, ErrUserBanned
	}

	// Update last login
	now := time.Now()
	_ = s.userRepo.SetLastLogin(ctx, user.ID, now)

	// Issue tokens
	return s.issueTokens(ctx, user)
}

// --- Token Management ---

func (s *authService) RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (*dto.AuthTokensResponse, error) {
	tokenHash := hashToken(req.RefreshToken)

	rt, err := s.userRepo.FindRefreshToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("finding refresh token: %w", err)
	}

	if !rt.IsValid() {
		return nil, ErrInvalidToken
	}

	// Revoke old token (rotation)
	_ = s.userRepo.RevokeRefreshToken(ctx, tokenHash)

	user, err := s.userRepo.FindByID(ctx, rt.UserID)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}

	if !user.IsActive() {
		return nil, ErrUserBanned
	}

	return s.issueTokens(ctx, user)
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := hashToken(refreshToken)
	return s.userRepo.RevokeRefreshToken(ctx, tokenHash)
}

// --- Profile ---

func (s *authService) GetProfile(ctx context.Context, userID string) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := toUserResponse(user)
	return &resp, nil
}

func (s *authService) UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("updating profile: %w", err)
	}

	resp := toUserResponse(user)
	return &resp, nil
}

// --- Helpers ---

// issueTokens mints a fresh access/refresh token pair and persists the refresh token.
func (s *authService) issueTokens(ctx context.Context, user *domain.User) (*dto.AuthTokensResponse, error) {
	// Access token
	accessToken, err := s.mintAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("minting access token: %w", err)
	}

	// Refresh token (opaque random string)
	rawRefresh, err := randomHex(32)
	if err != nil {
		return nil, fmt.Errorf("generating refresh token: %w", err)
	}

	rt := &domain.RefreshToken{
		ID:        newUUID(),
		UserID:    user.ID,
		Token:     hashToken(rawRefresh),
		ExpiresAt: time.Now().Add(s.cfg.JWT.RefreshExpiry),
		CreatedAt: time.Now(),
	}

	if err := s.userRepo.SaveRefreshToken(ctx, rt); err != nil {
		return nil, fmt.Errorf("saving refresh token: %w", err)
	}

	return &dto.AuthTokensResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		TokenType:    "Bearer",
		User:         toUserResponse(user),
	}, nil
}

// JWTClaims are the claims embedded in the access token.
type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (s *authService) mintAccessToken(user *domain.User) (string, error) {
	claims := JWTClaims{
		UserID: user.ID,
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.JWT.AccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.AccessSecret))
}

// ValidateAccessToken parses and validates a JWT access token.
func ValidateAccessToken(tokenStr, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// generateOTP returns a cryptographically random numeric OTP of the given length.
func generateOTP(length int) (string, error) {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", length, n), nil
}

// hashToken returns a SHA-256 hex hash of the token string for safe storage.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// randomHex generates a cryptographically random hex string of the given byte length.
func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// normalizePhone ensures the phone is stored consistently.
// For production replace this with a proper libphonenumber parse.
func normalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if !strings.HasPrefix(phone, "+") {
		phone = "+91" + phone // default to India; adjust for multi-country
	}
	return phone
}

// newUUID generates a simple UUID-like string using crypto/rand.
func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// toUserResponse converts a domain User to a safe DTO.
func toUserResponse(u *domain.User) dto.UserResponse {
	return dto.UserResponse{
		ID:         u.ID,
		Phone:      u.Phone,
		Name:       u.Name,
		Email:      u.Email,
		AvatarURL:  u.AvatarURL,
		Role:       string(u.Role),
		IsVerified: u.IsVerified,
	}
}

