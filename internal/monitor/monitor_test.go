package monitor

import (
	"testing"
	"time"

	"cronwatch/internal/config"
)

func makeConfig(jobs []config.Job) *config.Config {
	return &config.Config{
		CheckIntervalSec: 60,
		Jobs:             jobs,
	}
}

func TestRecordRun_SlowJob(t *testing.T) {
	cfg := makeConfig([]config.Job{
		{Name: "backup", IntervalSec: 3600, MaxDurationSec: 10},
	})
	m := New(cfg)

	m.RecordRun("backup", 20*time.Second)

	select {
	case alert := <-m.Alerts():
		if alert.Type != AlertSlow {
			t.Errorf("expected AlertSlow, got %s", alert.Type)
		}
		if alert.JobName != "backup" {
			t.Errorf("expected job name 'backup', got %s", alert.JobName)
		}
	default:
		t.Error("expected a slow alert but got none")
	}
}

func TestRecordRun_FastJob_NoAlert(t *testing.T) {
	cfg := makeConfig([]config.Job{
		{Name: "cleanup", IntervalSec: 3600, MaxDurationSec: 60},
	})
	m := New(cfg)

	m.RecordRun("cleanup", 5*time.Second)

	select {
	case alert := <-m.Alerts():
		t.Errorf("unexpected alert: %+v", alert)
	default:
		// pass
	}
}

func TestCheck_MissedJob(t *testing.T) {
	cfg := makeConfig([]config.Job{
		{Name: "report", IntervalSec: 60, MaxDurationSec: 30},
	})
	m := New(cfg)

	// Seed a last-seen time far in the past
	m.mu.Lock()
	m.states["report"] = &JobState{LastSeen: time.Now().Add(-5 * time.Minute)}
	m.mu.Unlock()

	m.Check(time.Now())

	select {
	case alert := <-m.Alerts():
		if alert.Type != AlertMissed {
			t.Errorf("expected AlertMissed, got %s", alert.Type)
		}
	default:
		t.Error("expected a missed alert but got none")
	}
}

func TestCheck_NoAlertWhenRecentRun(t *testing.T) {
	cfg := makeConfig([]config.Job{
		{Name: "ping", IntervalSec: 300, MaxDurationSec: 10},
	})
	m := New(cfg)

	m.mu.Lock()
	m.states["ping"] = &JobState{LastSeen: time.Now().Add(-30 * time.Second)}
	m.mu.Unlock()

	m.Check(time.Now())

	select {
	case alert := <-m.Alerts():
		t.Errorf("unexpected alert: %+v", alert)
	default:
		// pass
	}
}

func TestRecordRun_UnknownJob_NoAlert(t *testing.T) {
	cfg := makeConfig([]config.Job{})
	m := New(cfg)

	// Should not panic or emit an alert
	m.RecordRun("ghost", 1*time.Second)

	select {
	case alert := <-m.Alerts():
		t.Errorf("unexpected alert for unknown job: %+v", alert)
	default:
		// pass
	}
}
