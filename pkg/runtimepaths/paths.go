package runtimepaths

import (
	"os"
	"path/filepath"
	"strings"
)

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func HomeDir() string {
	if home := firstEnv("JANE_AI_HOME", "PICOCLAW_HOME"); home != "" {
		return home
	}
	userHome, _ := os.UserHomeDir()
	preferred := filepath.Join(userHome, ".jane-ai")
	if _, err := os.Stat(preferred); err == nil {
		return preferred
	}
	legacy := filepath.Join(userHome, ".picoclaw")
	if _, err := os.Stat(legacy); err == nil {
		return legacy
	}
	return preferred
}

func ConfigPath() string {
	if path := firstEnv("JANE_AI_CONFIG", "PICOCLAW_CONFIG"); path != "" {
		return path
	}
	return filepath.Join(HomeDir(), "config.json")
}

func BuiltinSkillsOverride() string {
	return firstEnv("JANE_AI_BUILTIN_SKILLS", "PICOCLAW_BUILTIN_SKILLS")
}
