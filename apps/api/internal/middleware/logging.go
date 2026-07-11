package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func RequestLogging(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return hlog.RequestIDHandler("reqId", "X-Request-Id")(
			hlog.MethodHandler("method")(
				hlog.URLHandler("url")(
					hlog.RemoteAddrHandler("remote")(
						hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
							hlog.FromRequest(r).Info().
								Int("status", status).
								Int("size", size).
								Dur("duration", duration).
								Msg("request")
						})(next),
					),
				),
			),
		)
	}
}
