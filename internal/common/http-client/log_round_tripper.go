package http_client

import (
	"currency-rate-app/internal/common/tracing"
	"log/slog"
	"net/http"
)

type LogRoundTripper struct {
	DefaultClient http.RoundTripper
}

func (i LogRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	slog.InfoContext(
		req.Context(),
		"HTTP Client Request",
		slog.String("method", req.Method),
		slog.String("host", req.Host),
		slog.String("url", req.URL.Path),
	)

	traceId, ok := tracing.TraceIDFromContext(req.Context())

	if ok {
		req.Header.Set("Trace-Id", traceId)
	}

	return i.DefaultClient.RoundTrip(req)
}
