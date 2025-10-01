package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func TestTracingMiddleware(t *testing.T) {
	handler := TracingMiddleware(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	got := rec.Header().Get(traceIdHeader)

	assert.NotNil(t, got, "expected trace-id header to be defined")
}
