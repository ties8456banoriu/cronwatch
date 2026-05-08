package monitor

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cronwatch/cronwatch/internal/alert"
	"github.com/cronwatch/cronwatch/internal/history"
)

// BurndownChecker periodically evaluates error-budget burndown for each
// configured job and fires an alert when the budget is exhausted.
type BurndownChecker struct {
	store         history.Store
	notifier      alert.Notifier
	window        time.Duration
	budgetPercent float64
	interval      time.Duration
}

// NewBurndownChecker creates a BurndownChecker.
func NewBurndownChecker(st history.Store, n alert.Notifier, window time.Duration, budgetPct float64, checkInterval time.Duration) *BurndownChecker {
	if checkInterval <= 0 {
		checkInterval = 10 * time.Minute
	}
	return &BurndownChecker{
		store:         st,
		notifier:      n,
		window:        window,
		budgetPercent: budgetPct,
		interval:      checkInterval,
	}
}

// Run starts the burndown check loop and blocks until ctx is cancelled.
func (b *BurndownChecker) Run(ctx context.Context, jobs []string) {
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, job := range jobs {
				b.checkJob(ctx, job)
			}
		}
	}
}

func (b *BurndownChecker) checkJob(ctx context.Context, job string) {
	res, err := history.ComputeBurndown(b.store, job, history.BurndownOptions{
		Window:        b.window,
		BudgetPercent: b.budgetPercent,
	})
	if err != nil {
		log.Printf("burndown check: %v", err)
		return
	}
	if !res.Exhausted {
		return
	}
	msg := fmt.Sprintf(
		"[cronwatch] error budget exhausted for job %q (budget=%.2f%%, exhausted at %s)",
		job, b.budgetPercent, res.ExhaustedAt.Format(time.RFC3339),
	)
	if err := b.notifier.Send(ctx, msg); err != nil {
		log.Printf("burndown check: send alert: %v", err)
	}
}
