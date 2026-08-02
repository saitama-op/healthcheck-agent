package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type URLConfig struct {
	URL            string `yaml:"url"`
	ExpectedStatus int    `yaml:"expected_status"`
}

type Config struct {
	Timeout               time.Duration `yaml:"timeout"`
	MinimumSuccessPercent float64       `yaml:"minimum_success_percent"`
	UserAgent             string        `yaml:"user_agent"`
	BindInterface         string        `yaml:"bind_interface"`
	DNSResolver           string        `yaml:"dns_resolver"` // NEW: e.g., "8.8.8.8" or "1.1.1.1:53"
	URLs                  []URLConfig   `yaml:"urls"`
}

// Load reads and parses the YAML configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg := &Config{
		Timeout:               3 * time.Second,
		MinimumSuccessPercent: 50.0,
		UserAgent:             "healthcheck-agent/1.0",
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if len(cfg.URLs) == 0 {
		return nil, fmt.Errorf("at least one URL must be configured")
	}

	return cfg, nil
}
