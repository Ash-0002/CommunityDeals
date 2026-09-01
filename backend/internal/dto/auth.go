package dto

// --- Request DTOs ---

// SendOTPRequest is the body for POST /auth/send-otp.
type SendOTPRequest struct {
	Phone string `json:"phone" binding:"required,min=10,max=15"`
}

// VerifyOTPRequest is the body for POST /auth/verify-otp.
type VerifyOTPRequest struct {
	Phone string `json:"phone" binding:"required,min=10,max=15"`
	OTP   string `json:"otp"   binding:"required,len=6"`
}

// RefreshTokenRequest is the body for POST /auth/refresh-token.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// UpdateProfileRequest is the body for PATCH /users/profile.
type UpdateProfileRequest struct {
	Name  string `json:"name"  binding:"omitempty,min=2,max=100"`
	Email string `json:"email" binding:"omitempty,email"`
}

// --- Response DTOs ---

// SendOTPResponse confirms that an OTP was dispatched.
type SendOTPResponse struct {
	Phone   string `json:"phone"`
	Message string `json:"message"`
	// ExpiresInSeconds tells the client when the OTP window closes.
	ExpiresInSeconds int `json:"expires_in_seconds"`
}

// AuthTokensResponse carries the JWT pair returned after a successful OTP verify.
type AuthTokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"` // "Bearer"
	User         UserResponse `json:"user"`
}

// UserResponse is a safe, serialisable view of a User domain object.
type UserResponse struct {
	ID         string `json:"id"`
	Phone      string `json:"phone"`
	Name       string `json:"name"`
	Email      string `json:"email,omitempty"`
	AvatarURL  string `json:"avatar_url,omitempty"`
	Role       string `json:"role"`
	IsVerified bool   `json:"is_verified"`
}
