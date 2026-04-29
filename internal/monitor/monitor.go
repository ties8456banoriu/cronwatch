package monitor

import (
	"log"
	"sync"
	"time"

	"cronwatch/internal/config"
)

// JobState tracks the last seen execution time and status of a cron job.
type JobState struct {
	LastSeen  time.Time
	Missed    bool
	Slow      bool
}

// Monitor watches configured cron jobs and emits alerts for missed or slow runs.
type Monitor struct {
	cfg    *config.Config
	states map[string]*JobState
	mu     sync.Mutex
	alerts chan Alert
}

// AlertType categorizes the kind of alert.
type AlertType string

const (
	AlertMissed AlertType = "missed"
	AlertSlow   AlertType = "slow"
)

// Alert represents a notification about a job anomaly.
type Alert struct {
	JobName  string
	Type     AlertType
	Message  string
	Occurred time.Time
}

// New creates a Monitor from the provided config.
func New(cfg *config.Config) *Monitor {
	return &Monitor{
		cfg:    cfg,
		states: make(map[string]*JobState),
		alerts: make(chan Alert, 64),
	}
}

// Alerts returns a read-only channel of emitted alerts.
func (m *Monitor) Alerts() <-chan Alert {
	return m.alerts
}

// RecordRun records a completed job run with its actual duration.
func (m *Monitor) RecordRun(jobName string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	job := m.findJob(jobName)
	if job == nil {
		log.Printf("[monitor] unknown job: %s", jobName)
		return
	}

	state := m.getOrCreateState(jobName)
	state.LastSeen = time.Now()
	state.Missed = false

	maxDuration := time.Duration(job.MaxDurationSec) * time.Second
	if maxDuration > 0 && duration > maxDuration {
		state.Slow = true
		m.emit(Alert{
			JobName:  jobName,
			Type:     AlertSlow,
			Message:  jobName + " exceeded max duration",
			Occurred: time.Now(),
		})
	} else {
		state.Slow = false
	}
}

// Check evaluates all jobs for missed runs based on their schedule interval.
func (m *Monitor) Check(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, job := range m.cfg.Jobs {
		state := m.getOrCreateState(job.Name)
		if state.LastSeen.IsZero() {
			continue
		}
		expected := time.Duration(job.IntervalSec) * time.Second
		if now.Sub(state.LastSeen) > expected && !state.Missed {
			state.Missed = true
			m.emit(Alert{
				JobName:  job.Name,
				Type:     AlertMissed,
				Message:  job.Name + " has not run within its expected interval",
				Occurred: now,
			})
		}
	}
}

func (m *Monitor) findJob(name string) *config.Job {
	for i := range m.cfg.Jobs {
		if m.cfg.Jobs[i].Name == name {
			return &m.cfg.Jobs[i]
		}
	}
	return nil
}

func (m *Monitor) getOrCreateState(name string) *JobState {
	if _, ok := m.states[name]; !ok {
		m.states[name] = &JobState{}
	}
	return m.states[name]
}

func (m *Monitor) emit(a Alert) {
	select {
	case m.alerts <- a:
	default:
		log.Printf("[monitor] alert channel full, dropping alert for %s", a.JobName)
	}
}
