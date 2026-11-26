package handlers

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

type galleryClaims struct {
	jwt.RegisteredClaims
	Slug string `json:"slug"`
}

func hasGalleryAccess(r *http.Request, slug string) bool {
	cookie, err := r.Cookie("gallery_access")
	if err != nil {
		return false
	}

	token, err := jwt.ParseWithClaims(cookie.Value, &galleryClaims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtSigningKey, nil
	})
	if err != nil {
		return false
	}

	claims, ok := token.Claims.(*galleryClaims)
	if !ok || claims.Slug != slug {
		return false
	}

	return true
}

func setGalleryCookie(w http.ResponseWriter, slug string, r *http.Request) {
	claims := &galleryClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Slug: slug,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(jwtSigningKey)
	if err != nil {
		log.Error().Err(err).Msg("failed to sign gallery JWT")
		return
	}

	// SameSite=None requires Secure. TLS terminates at the proxy (Fly.io /
	// Vercel), so r.TLS is nil there — trust X-Forwarded-Proto instead.
	secure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

	http.SetCookie(w, &http.Cookie{
		Name:     "gallery_access",
		Value:    signed,
		Path:     "/api",
		MaxAge:   3600,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   secure,
	})
}

var jwtSigningKey []byte

func SetJWTSecret(secret string) {
	jwtSigningKey = []byte(secret)
}
