package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

func TestHTTPSIndexReturnsServerErrorWhenTemplateParseFails(t *testing.T) {
	mudlog.SetupLogger(nil, "", "", false)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	repoRoot := filepath.Clean(filepath.Join(wd, "..", ".."))
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("os.Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("os.Chdir() cleanup error = %v", err)
		}
	})

	adminDir := t.TempDir()
	for _, name := range []string{"_header.html", "_footer.html"} {
		if err := os.WriteFile(filepath.Join(adminDir, name), []byte("test"), 0o600); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}
	}

	overridePath := filepath.Join(t.TempDir(), "config-overrides.yaml")
	overrideConfig := "FilePaths:\n  AdminHtml: \"" + strings.ReplaceAll(adminDir, "\\", "\\\\") + "\"\n"
	if err := os.WriteFile(overridePath, []byte(overrideConfig), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	oldConfigPath, hadConfigPath := os.LookupEnv("CONFIG_PATH")
	if err := os.Setenv("CONFIG_PATH", overridePath); err != nil {
		t.Fatalf("os.Setenv() error = %v", err)
	}
	if err := configs.ReloadConfig(); err != nil {
		t.Fatalf("configs.ReloadConfig() error = %v", err)
	}
	t.Cleanup(func() {
		if hadConfigPath {
			_ = os.Setenv("CONFIG_PATH", oldConfigPath)
		} else {
			_ = os.Unsetenv("CONFIG_PATH")
		}
		if err := configs.ReloadConfig(); err != nil {
			t.Fatalf("configs.ReloadConfig() cleanup error = %v", err)
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/admin/https/", nil)
	rec := httptest.NewRecorder()

	httpsIndex(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("httpsIndex() status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(rec.Body.String(), "Error parsing template files") {
		t.Fatalf("httpsIndex() body = %q, want parse failure message", rec.Body.String())
	}
}

func TestHTTPSIndexDisablesCaching(t *testing.T) {
	mudlog.SetupLogger(nil, "", "", false)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	repoRoot := filepath.Clean(filepath.Join(wd, "..", ".."))
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("os.Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("os.Chdir() cleanup error = %v", err)
		}
	})

	adminDir := t.TempDir()
	for name, contents := range map[string]string{
		"_header.html": "header",
		"https.html":   "nav={{len .NAV}}|{{.httpsStatus}}",
		"_footer.html": "footer",
	} {
		if err := os.WriteFile(filepath.Join(adminDir, name), []byte(contents), 0o600); err != nil {
			t.Fatalf("os.WriteFile() error = %v", err)
		}
	}

	overridePath := filepath.Join(t.TempDir(), "config-overrides.yaml")
	overrideConfig := "FilePaths:\n  AdminHtml: \"" + strings.ReplaceAll(adminDir, "\\", "\\\\") + "\"\n"
	if err := os.WriteFile(overridePath, []byte(overrideConfig), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	oldConfigPath, hadConfigPath := os.LookupEnv("CONFIG_PATH")
	if err := os.Setenv("CONFIG_PATH", overridePath); err != nil {
		t.Fatalf("os.Setenv() error = %v", err)
	}
	if err := configs.ReloadConfig(); err != nil {
		t.Fatalf("configs.ReloadConfig() error = %v", err)
	}
	t.Cleanup(func() {
		if hadConfigPath {
			_ = os.Setenv("CONFIG_PATH", oldConfigPath)
		} else {
			_ = os.Unsetenv("CONFIG_PATH")
		}
		if err := configs.ReloadConfig(); err != nil {
			t.Fatalf("configs.ReloadConfig() cleanup error = %v", err)
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/admin/https/", nil)
	rec := httptest.NewRecorder()

	httpsIndex(rec, req)

	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("httpsIndex() Cache-Control = %q, want %q", got, "no-store")
	}
	if strings.Contains(rec.Body.String(), "nav=0|") {
		t.Fatalf("httpsIndex() body = %q, want non-empty admin nav", rec.Body.String())
	}
}
