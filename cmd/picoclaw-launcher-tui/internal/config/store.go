package configstore

import (
	"errors"
	"path/filepath"

	picoclawconfig "jane/pkg/config"
	"jane/pkg/runtimepaths"
)

const (
	configFileName = "config.json"
)

func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

func ConfigDir() (string, error) {
	return runtimepaths.HomeDir(), nil
}

func Load() (*picoclawconfig.Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	return picoclawconfig.LoadConfig(path)
}

func Save(cfg *picoclawconfig.Config) error {
	if cfg == nil {
		return errors.New("config is nil")
	}
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	return picoclawconfig.SaveConfig(path, cfg)
}
