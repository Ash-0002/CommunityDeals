package middleware

import (
	"time"

	"github.com/community-platform/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLogger logs structured information about every HTTP request.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		if query != "" {
			path = path + "?" + query
		}

		log := logger.Get()
		statusCode := c.Writer.Status()

		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		// Add user_id if authenticated
		if userID := c.GetString(ContextUserID); userID != "" {
			fields = append(fields, zap.String("user_id", userID))
		}

		// Choose log level based on status code
		switch {
		case statusCode >= 500:
			log.Error("server error", fields...)
		case statusCode >= 400:
			log.Warn("client error", fields...)
		default:
			log.Info("request", fields...)
		}
	}
}
