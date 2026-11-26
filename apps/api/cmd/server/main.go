package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"xians/api/internal/config"
	"xians/api/internal/database"
	"xians/api/internal/handlers"
	"xians/api/internal/middleware"
	"xians/api/internal/models"
	"xians/api/internal/services"

	"xians/api/internal/storage"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Caller().
		Logger()

	for _, path := range []string{".env", "apps/api/.env"} {
		if _, err := os.Stat(path); err == nil {
			godotenv.Load(path)
			break
		}
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("invalid configuration")
	}

	log.Info().
		Str("port", cfg.Port).
		Int("maxFileMB", cfg.MaxFileSizeMB).
		Int("maxBatchMB", cfg.MaxBatchSizeMB).
		Msg("starting xians-api")

	dbConn, err := database.Connect(cfg.DatabaseURL, cfg.DBMaxOpenConns, cfg.DBMaxIdleConns)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer func() {
		sqlDB, _ := dbConn.DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()
	log.Info().Msg("connected to database")

	if cfg.AutoMigrate {
		if err := dbConn.DB.AutoMigrate(&models.Batch{}, &models.File{}, &models.RateLimit{}); err != nil {
			log.Fatal().Err(err).Msg("failed to run migrations")
		}
	}

	repo := models.NewBatchRepository(dbConn.DB)

	r2Client, err := storage.NewR2Client(
		cfg.R2AccountID,
		cfg.R2AccessKeyID,
		cfg.R2SecretAccessKey,
		cfg.R2BucketName,
		cfg.R2PublicHostname,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create R2 client")
	}

	uploadSvc := services.NewUploadService(cfg, repo, r2Client)
	gallerySvc := services.NewGalleryService(repo, r2Client)
	cleanupSvc := services.NewCleanupService(repo, r2Client)
	batchHandler := handlers.NewBatchHandler(uploadSvc, gallerySvc)
	cronHandler := handlers.NewCronHandler(cleanupSvc)

	handlers.SetJWTSecret(cfg.JWTSecret)
	handlers.SetCronSecret(cfg.CronSecret)

	r := chi.NewRouter()

	r.Use(chimw.RealIP)
	r.Use(middleware.Recover)
	r.Use(middleware.CORS(cfg.AllowedOrigins()))
	r.Use(chimw.RequestID)
	r.Use(chimw.Logger)

	initLimiter := middleware.NewRateLimiter(dbConn.DB, cfg.RateLimitInitPerMin, time.Minute)
	unlockLimiter := middleware.NewRateLimiter(dbConn.DB, cfg.RateLimitUnlockPerMin, time.Minute)
	reportLimiter := middleware.NewRateLimiter(dbConn.DB, cfg.RateLimitReportPerHour, time.Hour)

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", batchHandler.Health)

		// Vercel Cron triggers this (see vercel.json); replaces cmd/reaper
		// as the scheduled cleanup job.
		r.Get("/cron/cleanup", cronHandler.Cleanup)

		r.Route("/batches", func(r chi.Router) {
			r.With(middleware.RateLimit(initLimiter, middleware.ClientIP)).Post("/init", batchHandler.Init)

			r.Post("/{id}/complete", batchHandler.Complete)
			r.Post("/{id}/files/{fileId}/upload", batchHandler.UploadFile)
		})

		r.Route("/g/{slug}", func(r chi.Router) {
			r.Get("/", batchHandler.GetGallery)
			r.With(middleware.RateLimit(unlockLimiter, func(req *http.Request) string {
				return req.URL.Path + ":" + middleware.ClientIP(req)
			})).Post("/unlock", batchHandler.Unlock)
			r.Get("/files/{fileId}/view", batchHandler.ViewFile)
			r.Get("/files/{fileId}", batchHandler.DownloadFile)
			r.Get("/zip", batchHandler.DownloadZip)
			r.With(middleware.RateLimit(reportLimiter, middleware.ClientIP)).Post("/report", batchHandler.Report)
			r.Delete("/", batchHandler.Delete)
		})
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
		// Let the platform (Vercel/Fly) enforce overall request deadlines;
		// only guard against slow headers here so long zip streams and large
		// proxy uploads are not cut off mid-response.
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("forced shutdown")
	}
	log.Info().Msg("server stopped")
}
