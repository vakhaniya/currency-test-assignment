package middlewares

import (
	"log/slog"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		slog.InfoContext(
			r.Context(),
			"HTTP Server Request",
			slog.String("method", r.Method),
			slog.String("url", r.URL.Path),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
	})
}
