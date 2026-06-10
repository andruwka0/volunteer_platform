package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerHost      string        `yaml:"server_host"`
	ServerPort      int           `yaml:"server_host"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

func defaultConfig() *Config {
	return &Config{
		ServerHost:      "localhost",
		ServerPort:      8000,
		ReadTimeout:     5,
		WriteTimeout:    10,
		IdleTimeout:     60,
		ShutdownTimeout: 10,
	}
}

var ServerConfig *Config

// Load читает config.yaml
func Load() (*Config, error) {
	cfg := defaultConfig()

	data, err := os.ReadFile("config.yaml")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
