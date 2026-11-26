package middleware

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// RateLimiter enforces fixed-window limits backed by Postgres. State lives in
// the database instead of process memory so limits hold across Vercel
// instances and cold starts.
type RateLimiter struct {
	db        *gorm.DB
	rate      int
	perWindow time.Duration
}

func NewRateLimiter(db *gorm.DB, rate int, per time.Duration) *RateLimiter {
	return &RateLimiter{db: db, rate: rate, perWindow: per}
}

// Allow reports whether the request may proceed. Fails open if the database
// is unavailable so an outage does not take the API down.
func (rl *RateLimiter) Allow(ctx context.Context, key string) bool {
	window := time.Now().UTC().Truncate(rl.perWindow)

	var count int
	err := rl.db.WithContext(ctx).Raw(`
		INSERT INTO rate_limits ("key", "window_start", "count")
		VALUES (?, ?, 1)
		ON CONFLICT ("key", "window_start") DO UPDATE SET "count" = rate_limits."count" + 1
		RETURNING "count"
	`, key, window).Scan(&count).Error
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("rate limiter unavailable, allowing request")
		return true
	}

	return count <= rl.rate
}

// ClientIP returns the rate-limit key for a request: the IP only, without
// the ephemeral source port. chi's RealIP middleware has already replaced
// RemoteAddr with the leftmost X-Forwarded-For value when a proxy is present.
func ClientIP(req *http.Request) string {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return host
}

func RateLimit(limiter *RateLimiter, keyFunc func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)
			if !limiter.Allow(r.Context(), key) {
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
