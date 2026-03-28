package runtimepaths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHomeDirPrefersJaneAIByDefault(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if got := HomeDir(); got != filepath.Join(os.Getenv("HOME"), ".jane-ai") {
		t.Fatalf("HomeDir() = %q", got)
	}
}

func TestHomeDirFallsBackToLegacyDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	legacy := filepath.Join(home, ".picoclaw")
	if err := os.MkdirAll(legacy, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := HomeDir(); got != legacy {
		t.Fatalf("HomeDir() = %q, want %q", got, legacy)
	}
}

func TestConfigPathHonorsBothEnvNames(t *testing.T) {
	t.Setenv("JANE_AI_CONFIG", "/tmp/jane/config.json")
	if got := ConfigPath(); got != "/tmp/jane/config.json" {
		t.Fatalf("ConfigPath() = %q", got)
	}
	t.Setenv("JANE_AI_CONFIG", "")
	t.Setenv("PICOCLAW_CONFIG", "/tmp/pico/config.json")
	if got := ConfigPath(); got != "/tmp/pico/config.json" {
		t.Fatalf("ConfigPath() = %q", got)
	}
}
