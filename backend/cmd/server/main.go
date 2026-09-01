package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/community-platform/backend/internal/config"
	"github.com/community-platform/backend/internal/database"
	"github.com/community-platform/backend/internal/handler"
	"github.com/community-platform/backend/internal/repository"
	"github.com/community-platform/backend/internal/service"
	"github.com/community-platform/backend/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// ── 1. Load configuration ────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// ── 2. Initialise logger ─────────────────────────────────────────────────
	logger.Init(cfg.Env)
	defer logger.Sync()
	log := logger.Get()

	log.Info("starting community platform API",
		zap.String("env", cfg.Env),
		zap.String("port", cfg.Port),
	)

	// ── 3. Connect to PostgreSQL (with retry for Docker Compose startup) ──────
	log.Info("connecting to postgresql...")
	db, err := database.WithRetry(cfg.Database, 10)
	if err != nil {
		log.Fatal("failed to connect to postgresql", zap.Error(err))
	}
	defer db.Close()
	log.Info("postgresql connected")

	// ── 4. Connect to Redis ───────────────────────────────────────────────────
	log.Info("connecting to redis...")
	rdb, err := database.NewRedis(cfg.Redis)
	if err != nil {
		log.Fatal("failed to connect to redis", zap.Error(err))
	}
	defer rdb.Close()
	log.Info("redis connected")

	// ── 5. Wire dependencies (Repositories → Services → Handlers) ────────────

	// Repositories
	userRepo      := repository.NewUserRepository(db)
	communityRepo := repository.NewCommunityRepository(db)
	campaignRepo  := repository.NewCampaignRepository(db)

	// Services
	authSvc      := service.NewAuthService(userRepo, rdb, cfg)
	communitySvc := service.NewCommunityService(communityRepo, userRepo)
	campaignSvc  := service.NewCampaignService(campaignRepo, communityRepo)

	// Handlers
	authHandler      := handler.NewAuthHandler(authSvc)
	communityHandler := handler.NewCommunityHandler(communitySvc)
	campaignHandler  := handler.NewCampaignHandler(campaignSvc)

	// ── 6. Build router ───────────────────────────────────────────────────────
	router := handler.NewRouter(handler.Dependencies{
		Cfg:              cfg,
		AuthHandler:      authHandler,
		CommunityHandler: communityHandler,
		CampaignHandler:  campaignHandler,
	})

	// ── 7. Start HTTP server with graceful shutdown ───────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in a goroutine
	go func() {
		log.Info("server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	// Wait for interrupt or terminate signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	// Give in-flight requests up to 10 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", zap.Error(err))
	}

	log.Info("server stopped cleanly")
}
