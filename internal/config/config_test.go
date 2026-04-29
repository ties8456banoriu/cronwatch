package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "cronwatch-*.json")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	raw := `{
		"check_interval": 60000000000,
		"alert_webhook": "http://example.com/hook",
		"jobs": [
			{"name": "backup", "schedule": "0 2 * * *", "max_duration": 3600000000000, "alert_on_miss": true}
		]
	}`
	path := writeTempConfig(t, raw)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(cfg.Jobs))
	}
	if cfg.Jobs[0].Name != "backup" {
		t.Errorf("expected name 'backup', got %q", cfg.Jobs[0].Name)
	}
	if cfg.CheckInterval != 60*time.Second {
		t.Errorf("unexpected check_interval: %v", cfg.CheckInterval)
	}
}

func TestLoad_DefaultCheckInterval(t *testing.T) {
	raw := `{"jobs": [{"name": "sync", "schedule": "* * * * *"}]}`
	path := writeTempConfig(t, raw)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CheckInterval != 30*time.Second {
		t.Errorf("expected default 30s, got %v", cfg.CheckInterval)
	}
}

func TestLoad_MissingName(t *testing.T) {
	raw := `{"jobs": [{"schedule": "* * * * *"}]}`
	path := writeTempConfig(t, raw)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing name")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
