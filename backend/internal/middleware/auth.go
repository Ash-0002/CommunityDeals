package middleware

import (
	"strings"

	"github.com/community-platform/backend/internal/service"
	"github.com/community-platform/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

const (
	// ContextUserID is the key used to store the authenticated user's ID in the Gin context.
	ContextUserID = "user_id"
	// ContextUserRole is the key used to store the user's role.
	ContextUserRole = "user_role"
)

// Authenticate is a middleware that validates the JWT access token in the Authorization header.
// On success it injects user_id and user_role into the Gin context.
func Authenticate(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "MISSING_TOKEN", "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			response.Unauthorized(c, "INVALID_TOKEN_FORMAT", "Authorization header must be: Bearer <token>")
			c.Abort()
			return
		}

		tokenStr := parts[1]

		claims, err := service.ValidateAccessToken(tokenStr, jwtSecret)
		if err != nil {
			response.Unauthorized(c, "INVALID_TOKEN", "Token is invalid or expired")
			c.Abort()
			return
		}

		// Inject identity into context for downstream handlers
		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUserRole, claims.Role)

		c.Next()
	}
}

// RequireRole is a middleware that allows only users with one of the given roles.
// Must be used after Authenticate.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(c *gin.Context) {
		role, exists := c.Get(ContextUserRole)
		if !exists {
			response.Unauthorized(c, "MISSING_ROLE", "Unauthorized")
			c.Abort()
			return
		}

		if !allowed[role.(string)] {
			response.Forbidden(c, "INSUFFICIENT_ROLE", "You do not have permission to access this resource")
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID extracts the authenticated user ID from the Gin context.
// Panics if called outside an Authenticate-protected route.
func GetUserID(c *gin.Context) string {
	return c.GetString(ContextUserID)
}

// GetUserRole extracts the authenticated user role from the Gin context.
func GetUserRole(c *gin.Context) string {
	return c.GetString(ContextUserRole)
}
