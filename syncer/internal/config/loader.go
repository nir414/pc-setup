package config

import (
	"fmt"
	"os"

	toml "github.com/pelletier/go-toml/v2"
)

// Load reads a TOML configuration file.
func Load(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	if cfg.SyncData == nil {
		cfg.SyncData = map[string]Section{}
	}

	return &cfg, nil
}
