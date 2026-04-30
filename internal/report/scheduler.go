package report

import (
	"context"
	"log"
	"time"

	"github.com/example/cronwatch/internal/alert"
	"github.com/example/cronwatch/internal/history"
)

// Scheduler periodically builds and dispatches reports.
type Scheduler struct {
	Interval  time.Duration
	Stores    map[string]*history.Store
	Notifier  alert.Notifier
}

// NewScheduler creates a Scheduler with the given report interval.
func NewScheduler(interval time.Duration, stores map[string]*history.Store, n alert.Notifier) *Scheduler {
	return &Scheduler{
		Interval: interval,
		Stores:   stores,
		Notifier: n,
	}
}

// Run starts the periodic reporting loop and blocks until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.dispatch(ctx)
		case <-ctx.Done():
			log.Println("report scheduler stopped")
			return
		}
	}
}

func (s *Scheduler) dispatch(ctx context.Context) {
	summaries := make(map[string]history.Summary, len(s.Stores))
	for name, store := range s.Stores {
		records, err := store.All()
		if err != nil {
			log.Printf("report: failed to read history for %s: %v", name, err)
			continue
		}
		summaries[name] = history.Summarize(records)
	}
	r := Build(summaries)
	text := r.Format()
	if err := s.Notifier.Send(ctx, "CronWatch Status Report", text); err != nil {
		log.Printf("report: failed to send report: %v", err)
	}
}
