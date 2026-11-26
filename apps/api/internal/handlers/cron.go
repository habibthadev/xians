package handlers

import (
	"context"
	"crypto/hmac"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"xians/api/internal/services"

	"github.com/rs/zerolog/log"
)

var cronSecret string

// SetCronSecret configures the bearer token Vercel sends with cron requests
// (CRON_SECRET environment variable).
func SetCronSecret(secret string) {
	cronSecret = secret
}

// CronHandler exposes scheduled work (previously the standalone reaper
// binary) as an HTTP endpoint so Vercel Cron can trigger it.
type CronHandler struct {
	cleanupSvc *services.CleanupService
}

func NewCronHandler(cleanupSvc *services.CleanupService) *CronHandler {
	return &CronHandler{cleanupSvc: cleanupSvc}
}

// Cleanup deletes expired galleries and prunes rate-limit counters.
// GET /api/cron/cleanup
func (h *CronHandler) Cleanup(w http.ResponseWriter, r *http.Request) {
	if !cronAuthorized(r) {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Stay well under Vercel's function duration limit even for large sweeps.
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Minute)
	defer cancel()

	if err := h.cleanupSvc.CleanupExpired(ctx); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("cron cleanup failed")
		writeError(w, http.StatusInternalServerError, "cleanup failed")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func cronAuthorized(r *http.Request) bool {
	if cronSecret != "" {
		auth := r.Header.Get("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(auth, prefix) {
			return false
		}
		got := []byte(strings.TrimPrefix(auth, prefix))
		want := []byte(cronSecret)
		return hmac.Equal(got, want)
	}

	// Fallback when CRON_SECRET is unset: accept only requests Vercel
	// originates for cron schedules.
	return r.Header.Get("x-vercel-cron-schedule") != "" ||
		strings.HasPrefix(r.Header.Get("User-Agent"), "vercel-cron/")
}
