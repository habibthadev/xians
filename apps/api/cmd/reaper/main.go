package main

import (
	"context"
	"os"
	"time"

	"xians/api/internal/config"
	"xians/api/internal/database"
	"xians/api/internal/models"
	"xians/api/internal/services"
	"xians/api/internal/storage"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
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

	log.Info().Msg("starting reaper")

	dbConn, err := database.Connect(cfg.DatabaseURL, 2, 1)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer func() {
		sqlDB, _ := dbConn.DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

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

	cleanupSvc := services.NewCleanupService(repo, r2Client)

	if err := cleanupSvc.CleanupExpired(context.Background()); err != nil {
		log.Error().Err(err).Msg("cleanup failed")
		os.Exit(1)
	}

	log.Info().Msg("reaper completed")
}
