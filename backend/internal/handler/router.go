package handler

import (
	"net/http"

	"github.com/community-platform/backend/internal/config"
	"github.com/community-platform/backend/internal/middleware"
	"github.com/community-platform/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// Dependencies groups all handler dependencies for clean router setup.
type Dependencies struct {
	Cfg              *config.Config
	AuthHandler      *AuthHandler
	CommunityHandler *CommunityHandler
	CampaignHandler  *CampaignHandler
	// Future handlers:
	// VendorHandler    *VendorHandler
	// AdminHandler     *AdminHandler
}

// NewRouter creates the Gin engine with all routes registered.
func NewRouter(deps Dependencies) *gin.Engine {
	if deps.Cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New() // Use gin.New() so we control which middleware runs

	// Global middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(corsMiddleware())

	// Health check — no auth required, useful for load balancer probes
	r.GET("/health", healthCheck)
	r.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "pong"}) })

	// API v1
	v1 := r.Group("/api/v1")

	// ── Auth (public) ──────────────────────────────────────────────────────────
	auth := v1.Group("/auth")
	{
		auth.POST("/send-otp", deps.AuthHandler.SendOTP)
		auth.POST("/verify-otp", deps.AuthHandler.VerifyOTP)
		auth.POST("/refresh-token", deps.AuthHandler.RefreshToken)
		auth.POST("/logout", middleware.Authenticate(deps.Cfg.JWT.AccessSecret), deps.AuthHandler.Logout)
	}

	// ── Users (authenticated) ──────────────────────────────────────────────────
	users := v1.Group("/users", middleware.Authenticate(deps.Cfg.JWT.AccessSecret))
	{
		users.GET("/profile", deps.AuthHandler.GetProfile)
		users.PATCH("/profile", deps.AuthHandler.UpdateProfile)
		users.GET("/communities", deps.CommunityHandler.MyCommunities)
		// users.GET("/my-campaigns", ...)  ← added in Campaign milestone
		// users.GET("/orders", ...)        ← added in Orders milestone
	}

	// ── Communities (authenticated) ────────────────────────────────────────────
	communities := v1.Group("/communities", middleware.Authenticate(deps.Cfg.JWT.AccessSecret))
	{
		communities.POST("", deps.CommunityHandler.CreateCommunity)
		communities.GET("", deps.CommunityHandler.ListCommunities)
		communities.GET("/:id", deps.CommunityHandler.GetCommunity)
		communities.POST("/:id/join", deps.CommunityHandler.JoinCommunity)
		communities.POST("/:id/leave", deps.CommunityHandler.LeaveCommunity)
		communities.GET("/:id/members", deps.CommunityHandler.ListCommunityMembers)
		communities.POST("/:id/members/:user_id/approve", deps.CommunityHandler.ApproveMember)
		// communities.GET("/:id/campaigns", ...)  ← added in Campaign milestone
	}

	// ── Campaigns ──────────────────────────────────────────────────────────────
	// Public: shareable campaign link — no auth needed (web users clicking WhatsApp link)
	r.GET("/c/:slug", deps.CampaignHandler.GetCampaignBySlug)

	// Authenticated campaign routes
	campaigns := v1.Group("/campaigns", middleware.Authenticate(deps.Cfg.JWT.AccessSecret))
	{
		campaigns.POST("", deps.CampaignHandler.CreateCampaign)
		campaigns.GET("", deps.CampaignHandler.ListCampaigns)
		campaigns.GET("/:id", deps.CampaignHandler.GetCampaign)
		campaigns.POST("/:id/join", deps.CampaignHandler.JoinCampaign)
		campaigns.POST("/:id/leave", deps.CampaignHandler.LeaveCampaign)
		campaigns.GET("/:id/participants", deps.CampaignHandler.ListParticipants)
		campaigns.PATCH("/:id/status", deps.CampaignHandler.UpdateCampaignStatus)
	}

	// Community campaigns shortcut
	communities.GET("/:id/campaigns", deps.CampaignHandler.ListCampaigns)

	// ── Admin (platform admin only) ────────────────────────────────────────────
	// admin := v1.Group("/admin",
	//   middleware.Authenticate(deps.Cfg.JWT.AccessSecret),
	//   middleware.RequireRole("PLATFORM_ADMIN"),
	// )
	// Registered in Admin milestone

	return r
}

// healthCheck returns a 200 with basic service information.
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, response.APIResponse{
		Success: true,
		Message: "Service is healthy",
		Data: gin.H{
			"service": "community-platform-api",
			"version": "1.0.0",
		},
	})
}

// corsMiddleware adds CORS headers appropriate for the MVP.
// In production, restrict AllowOrigins to your actual domains.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
