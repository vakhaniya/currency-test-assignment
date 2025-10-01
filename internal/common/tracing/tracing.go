package tracing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type ctxKey string

const traceIDKey ctxKey = "TraceId"

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

func TraceIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(traceIDKey)
	if v == nil {
		return "", false
	}
	return v.(string), true
}

func NewTraceID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)

	return hex.EncodeToString(b)
}
