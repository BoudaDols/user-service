package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dolsom/user-service/internal/config"
	"github.com/dolsom/user-service/internal/database"
	"github.com/dolsom/user-service/internal/handler"
	kafkapkg "github.com/dolsom/user-service/internal/kafka"
	"github.com/dolsom/user-service/internal/middleware"
	"github.com/dolsom/user-service/internal/repository"
	"github.com/dolsom/user-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(os.Stdout)

	cfg := config.Load()

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// ── Database ──────────────────────────────────────────────
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}

	// ── Repositories ──────────────────────────────────────────
	profileRepo := repository.NewProfileRepository(db)
	preferenceRepo := repository.NewPreferenceRepository(db)
	activityRepo := repository.NewActivityRepository(db)

	// ── Kafka ─────────────────────────────────────────────────
	producer := kafkapkg.NewProducer(cfg.KafkaBroker)
	defer producer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer := kafkapkg.NewConsumer(cfg.KafkaBroker, profileRepo, activityRepo)
	consumer.Start(ctx)

	// ── Services ──────────────────────────────────────────────
	profileSvc := service.NewProfileService(profileRepo, activityRepo, producer)
	preferenceSvc := service.NewPreferenceService(preferenceRepo, activityRepo, producer)
	activitySvc := service.NewActivityService(activityRepo)

	// ── Handlers ──────────────────────────────────────────────
	profileH := handler.NewProfileHandler(profileSvc)
	preferenceH := handler.NewPreferenceHandler(preferenceSvc)
	activityH := handler.NewActivityHandler(activitySvc)

	// ── Router ────────────────────────────────────────────────
	r := gin.New()
	r.Use(middleware.RequestLogger())
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api", middleware.RequireUserID(), middleware.ActivityLogger(activityRepo))
	{
		profiles := api.Group("/profiles")
		{
			profiles.POST("", profileH.Create)
			profiles.GET("/:user_id", profileH.Get)
			profiles.PUT("/:user_id", profileH.Update)
			profiles.GET("/:user_id/preferences", preferenceH.GetAll)
			profiles.PUT("/:user_id/preferences", preferenceH.Upsert)
			profiles.DELETE("/:user_id/preferences/:key", preferenceH.Delete)
			profiles.GET("/:user_id/activity", activityH.GetByUserID)
		}
	}

	// ── Server ────────────────────────────────────────────────
	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	go func() {
		log.Info().Str("port", cfg.AppPort).Msg("user-service started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
}
