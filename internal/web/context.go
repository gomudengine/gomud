package web

import (
	"context"
	"net/http"
)

type contextKey int

const (
	internalCallerKey contextKey = iota
	testModeKey
)

// withInternalContext returns a copy of ctx marked as an internal request.
func withInternalContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, internalCallerKey, true)
}

// IsInternalRequest reports whether r was dispatched via InternalRequest rather
// than arriving over the network. Handlers can use this to skip audit logging,
// rate limiting, or other concerns that only apply to external callers.
func IsInternalRequest(r *http.Request) bool {
	v, _ := r.Context().Value(internalCallerKey).(bool)
	return v
}

// withTestModeContext returns a copy of ctx marked as a test-mode request.
func withTestModeContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, testModeKey, true)
}

// IsTestModeRequest reports whether r was sent with X-Test-Mode: true.
func IsTestModeRequest(r *http.Request) bool {
	v, _ := r.Context().Value(testModeKey).(bool)
	return v
}
