package internal

import (
	"jane/pkg/config"
	"jane/pkg/logger"
	"jane/pkg/runtimepaths"
)

const Logo = "🦞"

// GetPicoclawHome returns the picoclaw home directory.
// Priority: $PICOCLAW_HOME > ~/.picoclaw
func GetPicoclawHome() string {
	return runtimepaths.HomeDir()
}

func GetConfigPath() string {
	return runtimepaths.ConfigPath()
}

func LoadConfig() (*config.Config, error) {
	cfg, err := config.LoadConfig(GetConfigPath())
	if err == nil && cfg != nil {
		// Initialize the logger workspace directory for session replays
		logger.SetWorkspaceDir(cfg.WorkspacePath())
	}
	return cfg, err
}
