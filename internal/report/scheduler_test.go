package report_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/report"
)

// stubNotifier records calls to Send.
type stubNotifier struct {
	mu    sync.Mutex
	calls []string
}

func (s *stubNotifier) Send(_ context.Context, title, body string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, title)
	return nil
}

func (s *stubNotifier) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.calls)
}

func TestScheduler_SendsReport(t *testing.T) {
	n := &stubNotifier{}
	sched := report.NewScheduler(50*time.Millisecond, nil, n)

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Millisecond)
	defer cancel()

	go sched.Run(ctx)
	<-ctx.Done()

	if n.count() < 1 {
		t.Error("expected at least one report to be sent")
	}
}

func TestScheduler_StopsOnCancel(t *testing.T) {
	n := &stubNotifier{}
	sched := report.NewScheduler(1*time.Hour, nil, n)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		sched.Run(ctx)
		close(done)
	}()

	cancel()
	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Error("scheduler did not stop after context cancellation")
	}
}
