package middlewares

import (
	error_utils "currency-rate-app/internal/common/error-utils"
	http_server "currency-rate-app/internal/common/http-server"
	"log/slog"
	"net/http"
	"runtime/debug"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.ErrorContext(
					r.Context(),
					"RecoveryMiddleware",
					slog.Any("error", err),
					slog.String("stack", string(debug.Stack())),
				)

				http_server.SendErrorResponse(w, error_utils.ErrInternalServerError("InternalServerError"))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
