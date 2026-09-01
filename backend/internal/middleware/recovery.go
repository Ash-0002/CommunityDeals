package middleware

import (
	"net/http"

	"github.com/community-platform/backend/pkg/logger"
	"github.com/community-platform/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery catches any panics in handlers and returns a clean 500 response
// instead of crashing the server.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Get().Error("panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.APIResponse{
					Success: false,
					Error: &response.APIError{
						Code:    "INTERNAL_SERVER_ERROR",
						Message: "An unexpected error occurred. Our team has been notified.",
					},
				})
			}
		}()
		c.Next()
	}
}
