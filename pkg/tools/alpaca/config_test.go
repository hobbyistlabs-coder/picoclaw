package alpaca

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveConfigFallsBackToDotEnvAlpaca(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env.alpaca")
	content := "ALPACA_API_KEY=test-key\nALPACA_SECRET_KEY=test-secret\nALPACA_API_URL=https://paper-api.alpaca.markets\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	t.Setenv("JANE_AI_ALPACA_ENV_FILE", path)
	t.Setenv("ALPACA_API_KEY", "")
	t.Setenv("ALPACA_SECRET_KEY", "")
	t.Setenv("ALPACA_API_URL", "")

	key, secret, baseURL := resolveConfig("", "", "")
	if key != "test-key" || secret != "test-secret" || baseURL != "https://paper-api.alpaca.markets" {
		t.Fatalf("resolveConfig() = (%q, %q, %q)", key, secret, baseURL)
	}
}

func TestResolveConfigSearchesParentDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env.alpaca")
	content := "ALPACA_API_KEY=parent-key\nALPACA_SECRET_KEY=parent-secret\n"
	if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	nested := filepath.Join(tmpDir, "web", "backend")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(oldWD)
	if err := os.Chdir(nested); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Setenv("JANE_AI_ALPACA_ENV_FILE", "")
	t.Setenv("PICOCLAW_ALPACA_ENV_FILE", "")
	t.Setenv("ALPACA_API_KEY", "")
	t.Setenv("ALPACA_SECRET_KEY", "")

	key, secret, _ := resolveConfig("", "", "")
	if key != "parent-key" || secret != "parent-secret" {
		t.Fatalf("resolveConfig() = (%q, %q)", key, secret)
	}
}
