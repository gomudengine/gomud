package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

func TestAPIV1GetConfig_ReturnsData(t *testing.T) {
	mudlog.SetupLogger(nil, "", "", false)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	repoRoot := filepath.Clean(filepath.Join(wd, "..", ".."))
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("Chdir(%q): %v", repoRoot, err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	if err := configs.ReloadConfig(); err != nil {
		t.Fatalf("ReloadConfig: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/config", nil)
	rec := httptest.NewRecorder()

	apiV1GetConfig(rec, req)

	t.Logf("Status: %d", rec.Code)
	t.Logf("Body length: %d", rec.Body.Len())
	t.Logf("Body: %.500s", rec.Body.String())

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("response body is empty")
	}

	var result APIResponse[map[string]any]
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("json.Decode: %v", err)
	}
	if !result.Success {
		t.Fatalf("success = false, error = %q", result.Error)
	}
	if len(result.Data) == 0 {
		t.Fatal("data map is empty")
	}
	t.Logf("Keys in data: %d", len(result.Data))
}
