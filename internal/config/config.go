package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// JobConfig defines a single cron job to monitor.
type JobConfig struct {
	Name            string        `json:"name"`
	Schedule        string        `json:"schedule"`
	MaxDuration     time.Duration `json:"max_duration"`
	AlertOnMiss     bool          `json:"alert_on_miss"`
	AlertOnSlow     bool          `json:"alert_on_slow"`
	GracePeriod     time.Duration `json:"grace_period"`
}

// Config is the top-level configuration for cronwatch.
type Config struct {
	Jobs          []JobConfig   `json:"jobs"`
	CheckInterval time.Duration `json:"check_interval"`
	LogLevel      string        `json:"log_level"`
	AlertWebhook  string        `json:"alert_webhook"`
}

// Load reads and parses the config file at the given path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening config: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decoding config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = 30 * time.Second
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	for i, job := range c.Jobs {
		if job.Name == "" {
			return fmt.Errorf("job[%d]: name is required", i)
		}
		if job.Schedule == "" {
			return fmt.Errorf("job %q: schedule is required", job.Name)
		}
	}
	return nil
}
