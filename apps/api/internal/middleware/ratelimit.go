package middleware

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	mu        sync.Mutex
	buckets   map[string]*tokenBucket
	rate      int
	perWindow time.Duration
}

type tokenBucket struct {
	tokens    int
	lastCheck time.Time
}

func NewRateLimiter(rate int, per time.Duration) *RateLimiter {
	return &RateLimiter{
		buckets:   make(map[string]*tokenBucket),
		rate:      rate,
		perWindow: per,
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, exists := rl.buckets[key]
	if !exists {
		rl.buckets[key] = &tokenBucket{tokens: rl.rate - 1, lastCheck: time.Now()}
		return true
	}

	elapsed := time.Since(b.lastCheck)
	b.tokens += int(elapsed / rl.perWindow) * rl.rate
	if b.tokens > rl.rate {
		b.tokens = rl.rate
	}
	b.lastCheck = time.Now()

	if b.tokens <= 0 {
		return false
	}

	b.tokens--
	return true
}

func RateLimit(limiter *RateLimiter, keyFunc func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)
			if !limiter.Allow(key) {
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
