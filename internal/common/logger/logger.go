package logger

import (
	"context"
	"log/slog"
	"os"

	"currency-rate-app/internal/common/tracing"
)

func ConfigureLogger(debugMode bool) {
	if debugMode {
		slog.SetDefault(slog.New(&TraceHandler{h: slog.NewTextHandler(os.Stdout, nil)}))

		return
	}

	slog.SetDefault(slog.New(&TraceHandler{h: slog.NewJSONHandler(os.Stdout, nil)}))
}

type TraceHandler struct {
	h slog.Handler
}

func (t *TraceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return t.h.Enabled(ctx, level)
}

func (t *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	if traceID, ok := tracing.TraceIDFromContext(ctx); ok {
		r.AddAttrs(slog.String("traceId", traceID))
	}
	return t.h.Handle(ctx, r)
}

func (t *TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TraceHandler{h: t.h.WithAttrs(attrs)}
}

func (t *TraceHandler) WithGroup(name string) slog.Handler {
	return &TraceHandler{h: t.h.WithGroup(name)}
}
