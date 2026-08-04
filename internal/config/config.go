// Package config loads the service's runtime configuration from a YAML file.
//
// Only genuinely machine-level settings live here (the listen address).
// Everything about the *site* — base URL, analytics ids, contact routing —
// comes from the environment via package site, so nothing deployment-specific
// is ever committed.
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads and parses the YAML config at filename.
func Load(filename string) (Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}
	return Parse(data)
}

// Parse decodes YAML config content. The embedded default goes through here, so
// the binary needs no file on disk to start.
func Parse(data []byte) (Config, error) {
	cfg := Config{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Config is the top-level config document.
type Config struct {
	App Server `yaml:"app"`
}

// Server is one listener's settings.
type Server struct {
	Address string `yaml:"address"`
}
