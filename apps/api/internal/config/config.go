package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	MaxFileSizeMB      int
	MaxBatchSizeMB     int
	MaxFilesPerBatch   int
	DefaultTTLHours    int
	DatabaseURL        string
	R2AccountID        string
	R2AccessKeyID      string
	R2SecretAccessKey  string
	R2BucketName       string
	R2PublicHostname   string
	JWTSecret          string
	AllowedOrigin      string
	RateLimitInitPerMin   int
	RateLimitUnlockPerMin int
	RateLimitReportPerHour int
	Port               string
}

func Load() (*Config, error) {
	cfg := &Config{
		MaxFileSizeMB:        getEnvInt("MAX_FILE_SIZE_MB", 100),
		MaxBatchSizeMB:       getEnvInt("MAX_BATCH_SIZE_MB", 500),
		MaxFilesPerBatch:     getEnvInt("MAX_FILES_PER_BATCH", 20),
		DefaultTTLHours:      getEnvInt("DEFAULT_TTL_HOURS", 168),
		RateLimitInitPerMin:   getEnvInt("RATE_LIMIT_INIT_PER_MIN", 10),
		RateLimitUnlockPerMin: getEnvInt("RATE_LIMIT_UNLOCK_PER_MIN", 5),
		RateLimitReportPerHour: getEnvInt("RATE_LIMIT_REPORT_PER_HOUR", 3),
		Port:                 getEnv("PORT", "8080"),
	}

	required := map[string]string{
		"DATABASE_URL":         os.Getenv("DATABASE_URL"),
		"R2_ACCOUNT_ID":        os.Getenv("R2_ACCOUNT_ID"),
		"R2_ACCESS_KEY_ID":     os.Getenv("R2_ACCESS_KEY_ID"),
		"R2_SECRET_ACCESS_KEY": os.Getenv("R2_SECRET_ACCESS_KEY"),
		"R2_BUCKET_NAME":       os.Getenv("R2_BUCKET_NAME"),
		"R2_PUBLIC_HOSTNAME":   os.Getenv("R2_PUBLIC_HOSTNAME"),
		"JWT_SIGNING_SECRET":   os.Getenv("JWT_SIGNING_SECRET"),
		"ALLOWED_ORIGIN":       os.Getenv("ALLOWED_ORIGIN"),
	}

	for name, val := range required {
		if val == "" {
			return nil, fmt.Errorf("required environment variable %s is not set", name)
		}
	}

	cfg.DatabaseURL = required["DATABASE_URL"]
	cfg.R2AccountID = required["R2_ACCOUNT_ID"]
	cfg.R2AccessKeyID = required["R2_ACCESS_KEY_ID"]
	cfg.R2SecretAccessKey = required["R2_SECRET_ACCESS_KEY"]
	cfg.R2BucketName = required["R2_BUCKET_NAME"]
	cfg.R2PublicHostname = required["R2_PUBLIC_HOSTNAME"]
	cfg.JWTSecret = required["JWT_SIGNING_SECRET"]
	cfg.AllowedOrigin = required["ALLOWED_ORIGIN"]

	if cfg.MaxFileSizeMB <= 0 || cfg.MaxFileSizeMB > 500 {
		return nil, fmt.Errorf("MAX_FILE_SIZE_MB must be between 1 and 500")
	}
	if cfg.MaxBatchSizeMB <= 0 || cfg.MaxBatchSizeMB > 2000 {
		return nil, fmt.Errorf("MAX_BATCH_SIZE_MB must be between 1 and 2000")
	}
	if cfg.MaxFilesPerBatch <= 0 || cfg.MaxFilesPerBatch > 100 {
		return nil, fmt.Errorf("MAX_FILES_PER_BATCH must be between 1 and 100")
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SIGNING_SECRET must be at least 32 characters")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		n, err := strconv.Atoi(val)
		if err == nil {
			return n
		}
	}
	return fallback
}

func (c *Config) MaxFileSizeBytes() int64 {
	return int64(c.MaxFileSizeMB) * 1024 * 1024
}

func (c *Config) MaxBatchSizeBytes() int64 {
	return int64(c.MaxBatchSizeMB) * 1024 * 1024
}

func (c *Config) AllowedOrigins() []string {
	origins := strings.Split(c.AllowedOrigin, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}
	return origins
}

func (c *Config) R2Endpoint() string {
	return fmt.Sprintf("https://%s.r2.cloudflarestorage.com", c.R2AccountID)
}
