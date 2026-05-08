package monitor

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/yourorg/cronwatch/internal/history"
)

// captureNotifier records every message sent to it.
type captureNotifier struct {
	mu   sync.Mutex
	msgs []string
}

func (c *captureNotifier) Send(_ context.Context, msg string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.msgs = append(c.msgs, msg)
	return nil
}

func (c *captureNotifier) Messages() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	copy := make([]string, len(c.msgs))
	for i, m := range c.msgs {
		copy[i] = m
	}
	return copy
}

func populateAgingStore(t *testing.T, s *history.Store, job string, earlyMs, recentMs time.Duration) {
	t.Helper()
	now := time.Now()
	for i := 0; i < 10; i++ {
		_ = s.Add(history.Record{JobName: job, StartedAt: now.Add(-time.Duration(30-i) * 24 * time.Hour), Duration: earlyMs, Status: "success"})
	}
	for i := 0; i < 10; i++ {
		_ = s.Add(history.Record{JobName: job, StartedAt: now.Add(-time.Duration(10-i) * 24 * time.Hour), Duration: recentMs, Status: "success"})
	}
}

func TestAgingChecker_SendsAlert_WhenAging(t *testing.T) {
	s := history.New(t.TempDir() + "/ac.db")
	populateAgingStore(t, s, "slow-job", 100*time.Millisecond, 300*time.Millisecond)

	n := &captureNotifier{}
	ac := NewAgingChecker(s, n, 10, 20.0, time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	ac.Run(ctx, []string{"slow-job"})

	msgs := n.Messages()
	if len(msgs) == 0 {
		t.Fatal("expected at least one aging alert, got none")
	}
	for _, m := range msgs {
		if len(m) == 0 {
			t.Errorf("empty alert message")
		}
	}
}

func TestAgingChecker_NoAlert_WhenStable(t *testing.T) {
	s := history.New(t.TempDir() + "/ac2.db")
	populateAgingStore(t, s, "stable-job", 100*time.Millisecond, 105*time.Millisecond)

	n := &captureNotifier{}
	ac := NewAgingChecker(s, n, 10, 20.0, time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	ac.Run(ctx, []string{"stable-job"})

	if msgs := n.Messages(); len(msgs) > 0 {
		t.Errorf("expected no alert for stable job, got %d", len(msgs))
	}
}

func TestAgingChecker_SkipsSilently_WhenInsufficientData(t *testing.T) {
	s := history.New(t.TempDir() + "/ac3.db")
	// Only 3 records — not enough for window=10
	now := time.Now()
	for i := 0; i < 3; i++ {
		_ = s.Add(history.Record{JobName: "sparse", StartedAt: now.Add(-time.Duration(i) * time.Hour), Duration: 100 * time.Millisecond, Status: "success"})
	}

	n := &captureNotifier{}
	ac := NewAgingChecker(s, n, 10, 20.0, time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	ac.Run(ctx, []string{"sparse"})

	if msgs := n.Messages(); len(msgs) > 0 {
		t.Errorf("expected no alert for sparse data, got %d", len(msgs))
	}
}
