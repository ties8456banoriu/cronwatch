package monitor

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yourorg/cronwatch/internal/alert"
	"github.com/yourorg/cronwatch/internal/history"
)

// AgingChecker periodically checks all configured jobs for long-term
// performance degradation using history.DetectAging.
type AgingChecker struct {
	store    *history.Store
	notifier alert.Notifier
	window   int
	threshold float64
	interval  time.Duration
}

// NewAgingChecker creates an AgingChecker.
// window is the number of samples per half-window; threshold is the percentage
// increase that marks a job as aging; interval controls how often checks run.
func NewAgingChecker(s *history.Store, n alert.Notifier, window int, threshold float64, interval time.Duration) *AgingChecker {
	if window <= 0 {
		window = 10
	}
	if threshold <= 0 {
		threshold = 20.0
	}
	if interval <= 0 {
		interval = time.Hour
	}
	return &AgingChecker{
		store:     s,
		notifier:  n,
		window:    window,
		threshold: threshold,
		interval:  interval,
	}
}

// Run starts the aging check loop and blocks until ctx is cancelled.
func (ac *AgingChecker) Run(ctx context.Context, jobNames []string) {
	ticker := time.NewTicker(ac.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ac.checkAll(ctx, jobNames)
		}
	}
}

func (ac *AgingChecker) checkAll(ctx context.Context, jobNames []string) {
	for _, name := range jobNames {
		res, err := history.DetectAging(ac.store, name, history.AgingOptions{
			WindowSize:   ac.window,
			ThresholdPct: ac.threshold,
		})
		if err != nil {
			// Not enough data yet — skip silently.
			continue
		}
		if res.IsAging {
			msg := fmt.Sprintf(
				"[cronwatch] aging detected for %q: recent avg %.0f ms vs early avg %.0f ms (+%.1f%%)",
				res.JobName, res.RecentAvgMs, res.EarlyAvgMs, res.DeltaPct,
			)
			if err := ac.notifier.Send(ctx, msg); err != nil {
				log.Printf("aging_check: failed to send alert for %q: %v", name, err)
			}
		}
	}
}
