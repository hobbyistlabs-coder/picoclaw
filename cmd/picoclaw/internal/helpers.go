package internal

import (
	"fmt"
	"os"

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
	path := GetConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found at %s", path)
	}
	return config.LoadConfig(path)
}
