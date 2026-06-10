package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// rawConfig — для парсинга YAML (числа)
type rawConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	ReadTimeout     int    `yaml:"read_timeout"`     // секунды
	WriteTimeout    int    `yaml:"write_timeout"`    // секунды
	IdleTimeout     int    `yaml:"idle_timeout"`     // секунды
	ShutdownTimeout int    `yaml:"shutdown_timeout"` // секунды
	WorkerInterval  int    `yaml:"worker_interval"`  // минуты
}

// Config — итоговая структура (с Duration)
type Config struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	WorkerInterval  time.Duration
}

func Load() (*Config, error) {
	raw := &rawConfig{
		Host:            "localhost",
		Port:            8080,
		ReadTimeout:     5,
		WriteTimeout:    10,
		IdleTimeout:     60,
		ShutdownTimeout: 10,
		WorkerInterval:  1,
	}

	configPaths := []string{
		"config.yaml",
		"backend/config.yaml",
		"../config.yaml",
	}

	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		configPaths = append([]string{envPath}, configPaths...)
	}

	var data []byte
	var err error
	var foundPath string

	for _, path := range configPaths {
		data, err = os.ReadFile(path)
		if err == nil {
			foundPath = path
			break
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("ошибка чтения %s: %w", path, err)
		}
	}

	if data != nil {
		if err := yaml.Unmarshal(data, raw); err != nil {
			return nil, fmt.Errorf("ошибка парсинга %s: %w", foundPath, err)
		}
	}

	return &Config{
		Host:            raw.Host,
		Port:            raw.Port,
		ReadTimeout:     time.Duration(raw.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(raw.WriteTimeout) * time.Second,
		IdleTimeout:     time.Duration(raw.IdleTimeout) * time.Second,
		ShutdownTimeout: time.Duration(raw.ShutdownTimeout) * time.Second,
		WorkerInterval:  time.Duration(raw.WorkerInterval) * time.Minute,
	}, nil
}
