package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "driftwatch-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTempConfig(t, `
poll_interval: 30s
log_level: debug
providers:
  - name: prod-aws
    type: aws
    region: us-east-1
alerts:
  log_only: true
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollInterval != 30*time.Second {
		t.Errorf("expected poll_interval 30s, got %s", cfg.PollInterval)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected log_level debug, got %s", cfg.LogLevel)
	}
	if len(cfg.Providers) != 1 || cfg.Providers[0].Name != "prod-aws" {
		t.Errorf("unexpected providers: %+v", cfg.Providers)
	}
	if !cfg.Alerts.LogOnly {
		t.Error("expected alerts.log_only to be true")
	}
}

func TestLoad_Defaults(t *testing.T) {
	path := writeTempConfig(t, `
providers:
  - name: staging
    type: gcp
    region: us-central1
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PollInterval != 60*time.Second {
		t.Errorf("expected default poll_interval 60s, got %s", cfg.PollInterval)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected default log_level info, got %s", cfg.LogLevel)
	}
}

func TestLoad_InvalidProviderType(t *testing.T) {
	path := writeTempConfig(t, `
providers:
  - name: bad
    type: digitalocean
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for unsupported provider type")
	}
}

func TestLoad_PollIntervalTooShort(t *testing.T) {
	path := writeTempConfig(t, `
poll_interval: 500ms
providers:
  - name: x
    type: aws
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for poll_interval < 1s")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/driftwatch.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
