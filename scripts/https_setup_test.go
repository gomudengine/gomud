package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const sampleHTTPSConfig = `FilePaths:
  DataFiles: _datafiles/world/default
  HttpsCertFile: "server.crt"
  HttpsKeyFile: "server.key"
Network:
  HttpPort: 8080
  HttpsPort: 8443
  HttpsRedirect: false
`

func TestHTTPSSetupManualModePrintsOverrideSnippet(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)

	input := strings.Join([]string{
		"1",
		"",
		"",
		"",
		"",
		"",
		"2",
		"",
	}, "\n")

	output := runHTTPSSetup(t, configPath, input, nil)
	if !strings.Contains(output, "Override target: _datafiles/world/default/config-overrides.yaml") {
		t.Fatalf("https-setup output did not show override target:\n%s", output)
	}
	if !strings.Contains(output, "Save the following override snippet") {
		t.Fatalf("https-setup output did not switch to snippet mode:\n%s", output)
	}
	if !strings.Contains(output, "HttpsCertFile: 'server.crt'") {
		t.Fatalf("https-setup output did not include cert override:\n%s", output)
	}

	updated := readHTTPSSetupConfig(t, configPath)
	if updated != sampleHTTPSConfig {
		t.Fatalf("https-setup changed bundled config unexpectedly:\n%s", updated)
	}
}

func TestHTTPSSetupHTTPOnlySnippetClearsHTTPS(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)

	input := strings.Join([]string{
		"2",
		"8080",
		"2",
		"",
	}, "\n")

	output := runHTTPSSetup(t, configPath, input, nil)
	if !strings.Contains(output, "HttpsCertFile: ''") {
		t.Fatalf("https-setup output did not clear cert override:\n%s", output)
	}
	if !strings.Contains(output, "HttpsPort: 0") {
		t.Fatalf("https-setup output did not disable HTTPS in snippet:\n%s", output)
	}
}

func TestHTTPSSetupUsesConfigPathAsOverrideTarget(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)
	overridePath := filepath.Join(t.TempDir(), "custom-overrides.yaml")

	input := strings.Join([]string{
		"2",
		"8080",
		"2",
		"",
	}, "\n")

	output := runHTTPSSetup(t, configPath, input, map[string]string{
		"CONFIG_PATH": overridePath,
	})
	if !strings.Contains(output, "Override target: "+overridePath) {
		t.Fatalf("https-setup output did not prefer CONFIG_PATH:\n%s", output)
	}
}

func TestHTTPSSetupUsesExistingOverrideValuesAsDefaults(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)
	overridePath := filepath.Join(t.TempDir(), "custom-overrides.yaml")
	overrideConfig := `FilePaths:
  HttpsCertFile: 'existing.crt'
  HttpsKeyFile: 'existing.key'
Network:
  HttpPort: 9080
  HttpsPort: 9443
  HttpsRedirect: true
`
	if err := os.WriteFile(overridePath, []byte(overrideConfig), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	input := strings.Join([]string{
		"1",
		"",
		"",
		"",
		"",
		"",
		"2",
		"",
	}, "\n")

	output := runHTTPSSetup(t, configPath, input, map[string]string{
		"CONFIG_PATH": overridePath,
	})
	if !strings.Contains(output, "HttpsCertFile: existing.crt") {
		t.Fatalf("https-setup output did not preserve override cert default:\n%s", output)
	}
	if !strings.Contains(output, "HttpPort: 9080") {
		t.Fatalf("https-setup output did not preserve override HTTP port default:\n%s", output)
	}
	if !strings.Contains(output, "HttpsPort: 9443") {
		t.Fatalf("https-setup output did not preserve override HTTPS port default:\n%s", output)
	}
	if !strings.Contains(output, "HttpsRedirect: true") {
		t.Fatalf("https-setup output did not preserve override redirect default:\n%s", output)
	}
}

func TestHTTPSSetupSnippetKeepsWindowsPathsYAMLSafe(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)

	input := strings.Join([]string{
		"1",
		`C:\certs\server.crt`,
		`C:\certs\server.key`,
		"",
		"",
		"",
		"2",
		"",
	}, "\n")

	output := runHTTPSSetup(t, configPath, input, nil)
	if !strings.Contains(output, "HttpsCertFile: 'C:\\certs\\server.crt'") {
		t.Fatalf("https-setup output did not emit YAML-safe Windows cert path:\n%s", output)
	}
	if !strings.Contains(output, "HttpsKeyFile: 'C:\\certs\\server.key'") {
		t.Fatalf("https-setup output did not emit YAML-safe Windows key path:\n%s", output)
	}
}

func TestHTTPSSetupIgnoresConfigPathWhenItTargetsBundledConfig(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)

	input := strings.Join([]string{
		"2",
		"8080",
		"2",
		"",
	}, "\n")

	output := runHTTPSSetup(t, configPath, input, map[string]string{
		"CONFIG_PATH": configPath,
	})
	if !strings.Contains(output, "Ignoring CONFIG_PATH="+configPath) {
		t.Fatalf("https-setup output did not warn about bundled config target:\n%s", output)
	}
	if !strings.Contains(output, "Override target: _datafiles/world/default/config-overrides.yaml") {
		t.Fatalf("https-setup output did not fall back to config-overrides target:\n%s", output)
	}
}

func TestHTTPSSetupCanPatchRunningServer(t *testing.T) {
	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)
	curlDir := t.TempDir()
	curlLog := filepath.Join(curlDir, "curl.log")
	curlStub := filepath.Join(curlDir, "curl-stub.sh")
	curlScript := "#!/usr/bin/env sh\nprintf '%s\\n' \"$@\" >\"" + curlLog + "\"\nexit 0\n"
	if err := os.WriteFile(curlStub, []byte(curlScript), 0o755); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	input := strings.Join([]string{
		"1",
		"",
		"",
		"",
		"",
		"",
		"1",
		"http://localhost/",
		"admin",
		"password",
		"",
	}, "\n")

	output := runHTTPSSetup(t, configPath, input, map[string]string{
		"CURL_BIN": curlStub,
	})
	if !strings.Contains(output, "HTTPS setup applied through the admin API.") {
		t.Fatalf("https-setup output did not confirm API patch:\n%s", output)
	}

	logBytes, err := os.ReadFile(curlLog)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}
	curlArgs := string(logBytes)
	if !strings.Contains(curlArgs, "http://localhost/admin/api/v1/config") {
		t.Fatalf("curl stub did not receive config endpoint:\n%s", curlArgs)
	}
	if !strings.Contains(curlArgs, "FilePaths.HttpsCertFile") {
		t.Fatalf("curl stub did not receive config payload:\n%s", curlArgs)
	}

	updated := readHTTPSSetupConfig(t, configPath)
	if updated != sampleHTTPSConfig {
		t.Fatalf("https-setup changed bundled config unexpectedly:\n%s", updated)
	}
}

func TestHTTPSSetupPreservesBundledConfigPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file modes are not stable on Windows")
	}

	configPath := writeHTTPSSetupTempConfig(t, sampleHTTPSConfig)
	if err := os.Chmod(configPath, 0o640); err != nil {
		t.Fatalf("os.Chmod() error = %v", err)
	}

	input := strings.Join([]string{
		"2",
		"8080",
		"2",
		"",
	}, "\n")

	runHTTPSSetup(t, configPath, input, nil)

	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o640); got != want {
		t.Fatalf("config mode = %o, want %o", got, want)
	}
}

func writeHTTPSSetupTempConfig(t *testing.T, content string) string {
	t.Helper()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	return configPath
}

func readHTTPSSetupConfig(t *testing.T, configPath string) string {
	t.Helper()

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}

	return string(data)
}

func runHTTPSSetup(t *testing.T, configPath string, input string, extraEnv map[string]string) string {
	t.Helper()

	scriptDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}

	tempDir := t.TempDir()
	cmd := exec.Command("sh", "./https-setup.sh")
	cmd.Dir = scriptDir
	cmd.Env = append(os.Environ(),
		"CONFIG_FILE="+configPath,
		"TMPDIR="+tempDir,
	)
	for key, value := range extraEnv {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("https-setup.sh failed: %v\n%s", err, output)
	}

	return string(output)
}
