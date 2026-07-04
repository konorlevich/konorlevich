package config

import (
	_ "net/http/pprof"
	"os"

	"github.com/stretchr/testify/assert/yaml"
)

// Load configuration from yaml file
func Load(filename string) (Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}
	cfg := Config{}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

type Config struct {
	App   Server `yaml:"app"`
	Pprof Server `yaml:"pprof"`
}

type Server struct {
	Address string `yaml:"address"`
}
