package internal

import (
	"jane/pkg/config"
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
	return config.LoadConfig(GetConfigPath())
}
