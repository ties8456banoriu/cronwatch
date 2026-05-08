package monitor

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/history"
)

type capturingNotifier struct {
	mu       sync.Mutex
	messages []string
}

func (c *capturingNotifier) Send(_ context.Context, msg string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = append(c.messages, msg)
	return nil
}

func (c *capturingNotifier) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.messages)
}

func populateBurndownStore(t *testing.T, st history.Store, job string, successes, failures int) {
	t.Helper()
	for i := 0; i < successes; i++ {
		_ = st.Add(history.Record{JobName: job, StartedAt: time.Now().Add(-time.Duration(i+failures+1) * time.Minute), Duration: time.Second, Success: true})
	}
	for i := 0; i < failures; i++ {
		_ = st.Add(history.Record{JobName: job, StartedAt: time.Now().Add(-time.Duration(i+1) * time.Minute), Duration: time.Second, Success: false})
	}
}

func TestBurndownChecker_SendsAlert_WhenExhausted(t *testing.T) {
	dir := t.TempDir()
	st := history.New(dir + "/burndown.db")
	// 5 failures, 0 successes → 100% failure rate, budget 1%
	populateBurndownStore(t, st, "job1", 0, 5)

	n := &capturingNotifier{}
	checker := NewBurndownChecker(st, n, 24*time.Hour, 1.0, 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	checker.Run(ctx, []string{"job1"})

	if n.count() == 0 {
		t.Error("expected at least one alert to be sent")
	}
}

func TestBurndownChecker_NoAlert_WhenHealthy(t *testing.T) {
	dir := t.TempDir()
	st := history.New(dir + "/burndown.db")
	// all successes → 0% failure rate
	populateBurndownStore(t, st, "job2", 10, 0)

	n := &capturingNotifier{}
	checker := NewBurndownChecker(st, n, 24*time.Hour, 1.0, 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	checker.Run(ctx, []string{"job2"})

	if n.count() != 0 {
		t.Errorf("expected no alerts, got %d", n.count())
	}
}

func TestBurndownChecker_StopsOnCancel(t *testing.T) {
	dir := t.TempDir()
	st := history.New(dir + "/burndown.db")
	n := &capturingNotifier{}
	checker := NewBurndownChecker(st, n, 24*time.Hour, 1.0, 10*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		checker.Run(ctx, []string{})
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("Run did not stop after context cancellation")
	}
}
