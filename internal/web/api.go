package web

import (
	"encoding/json"
	"net/http"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

const testModeHeader = "X-Test-Mode"

// APIResponse is the generic envelope used by every API endpoint.
type APIResponse[T any] struct {
	Success  bool   `json:"success"`
	Data     T      `json:"data,omitempty"`
	Error    string `json:"error,omitempty"`
	TestMode bool   `json:"test_mode,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		mudlog.Error("API", "action", "writeJSON", "error", err)
	}
}

func writeAPIError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, APIResponse[struct{}]{
		Success: false,
		Error:   message,
	})
}

// RunInTestMode wraps a handler so that when the request carries
// "X-Test-Mode: true", the current config overrides are snapshotted before
// the handler runs and restored unconditionally afterwards. The response
// carries an "X-Test-Mode: true" header to confirm the mode was active.
func RunInTestMode(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(testModeHeader) != "true" {
			next(w, r)
			return
		}

		// Snapshot current overrides before the handler mutates anything.
		snapshot := configs.Flatten(configs.GetOverrides())

		// Mark the request context so handlers can inspect it if needed.
		r = r.WithContext(withTestModeContext(r.Context()))

		w.Header().Set(testModeHeader, "true")
		next(w, r)

		// Restore the snapshot regardless of what the handler did.
		if err := configs.RestoreOverrides(snapshot); err != nil {
			mudlog.Error("API", "action", "RunInTestMode restore", "error", err)
		}
	}
}
