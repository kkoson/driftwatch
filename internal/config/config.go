package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level driftwatch daemon configuration.
type Config struct {
	PollInterval time.Duration `yaml:"poll_interval"`
	LogLevel     string        `yaml:"log_level"`
	Providers    []Provider    `yaml:"providers"`
	Alerts       AlertConfig   `yaml:"alerts"`
}

// Provider represents a single cloud provider target to monitor.
type Provider struct {
	Name    string            `yaml:"name"`
	Type    string            `yaml:"type"` // aws | gcp | azure
	Region  string            `yaml:"region"`
	Options map[string]string `yaml:"options"`
}

// AlertConfig configures how drift alerts are emitted.
type AlertConfig struct {
	SlackWebhook string `yaml:"slack_webhook"`
	Email        string `yaml:"email"`
	LogOnly      bool   `yaml:"log_only"`
}

// Load reads and parses a YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: reading file %q: %w", path, err)
	}

	cfg := &Config{
		PollInterval: 60 * time.Second,
		LogLevel:     "info",
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("config: parsing yaml: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: validation failed: %w", err)
	}

	return cfg, nil
}

// validate checks that required fields are present and values are sane.
func (c *Config) validate() error {
	if c.PollInterval < time.Second {
		return fmt.Errorf("poll_interval must be at least 1s, got %s", c.PollInterval)
	}
	for i, p := range c.Providers {
		if p.Name == "" {
			return fmt.Errorf("provider[%d]: name is required", i)
		}
		switch p.Type {
		case "aws", "gcp", "azure":
		default:
			return fmt.Errorf("provider[%d] %q: unsupported type %q", i, p.Name, p.Type)
		}
	}
	return nil
}
