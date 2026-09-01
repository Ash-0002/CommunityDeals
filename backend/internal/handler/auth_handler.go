package handler

import (
	"errors"
	"net/http"

	"github.com/community-platform/backend/internal/dto"
	"github.com/community-platform/backend/internal/middleware"
	"github.com/community-platform/backend/internal/service"
	"github.com/community-platform/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles all authentication-related HTTP endpoints.
type AuthHandler struct {
	authSvc service.AuthService
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(authSvc service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// SendOTP godoc
// POST /auth/send-otp
// Sends a one-time password to the given phone number.
func (h *AuthHandler) SendOTP(c *gin.Context) {
	var req dto.SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "VALIDATION_ERROR", "Invalid request body", err.Error())
		return
	}

	resp, err := h.authSvc.SendOTP(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRateLimited):
			response.TooManyRequests(c, "Too many OTP requests. Please wait before requesting again.")
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, resp.Message, resp)
}

// VerifyOTP godoc
// POST /auth/verify-otp
// Validates the OTP and returns JWT tokens. Creates account on first login.
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req dto.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "VALIDATION_ERROR", "Invalid request body", err.Error())
		return
	}

	resp, err := h.authSvc.VerifyOTP(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOTP):
			response.BadRequest(c, "INVALID_OTP", "OTP is invalid or has expired", nil)
		case errors.Is(err, service.ErrUserBanned):
			response.Forbidden(c, "ACCOUNT_BANNED", "Your account has been suspended. Please contact support.")
		default:
			response.InternalError(c)
		}
		return
	}

	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Login successful",
		Data:    resp,
	})
}

// RefreshToken godoc
// POST /auth/refresh-token
// Issues a new access token using a valid refresh token (token rotation).
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "VALIDATION_ERROR", "Invalid request body", err.Error())
		return
	}

	resp, err := h.authSvc.RefreshToken(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidToken):
			response.Unauthorized(c, "INVALID_REFRESH_TOKEN", "Refresh token is invalid or expired")
		case errors.Is(err, service.ErrUserBanned):
			response.Forbidden(c, "ACCOUNT_BANNED", "Your account has been suspended.")
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, "Token refreshed", resp)
}

// Logout godoc
// POST /auth/logout
// Revokes the provided refresh token. Requires authentication.
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "VALIDATION_ERROR", "Invalid request body", err.Error())
		return
	}

	if err := h.authSvc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "Logged out successfully", nil)
}

// GetProfile godoc
// GET /users/profile
// Returns the authenticated user's profile.
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	prof, err := h.authSvc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "Profile fetched", prof)
}

// UpdateProfile godoc
// PATCH /users/profile
// Updates the authenticated user's name and/or email.
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "VALIDATION_ERROR", "Invalid request body", err.Error())
		return
	}

	prof, err := h.authSvc.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "Profile updated", prof)
}
