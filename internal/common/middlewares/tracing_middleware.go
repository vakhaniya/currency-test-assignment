package middlewares

import (
	"net/http"

	"currency-rate-app/internal/common/tracing"
)

const traceIdHeader = "Trace-Id"

func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceId := r.Header.Get(traceIdHeader)

		if traceId == "" {
			traceId = tracing.NewTraceID()
		}

		w.Header().Set(traceIdHeader, traceId)

		ctx := tracing.WithTraceID(r.Context(), traceId)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
