package history

import (
	"testing"
	"time"
)

func addRecoveryRecord(t *testing.T, st *Store, job, status string, at time.Time) {
	t.Helper()
	err := st.Add(Record{
		JobName:   job,
		Status:    status,
		StartedAt: at,
		Duration:  200 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("addRecoveryRecord: %v", err)
	}
}

func TestDetectRecovery_IsRecovered(t *testing.T) {
	st := New(t.TempDir() + "/r.json")
	now := time.Now()
	for i := 0; i < 3; i++ {
		addRecoveryRecord(t, st, "job1", "failure", now.Add(time.Duration(i)*time.Minute))
	}
	for i := 3; i < 6; i++ {
		addRecoveryRecord(t, st, "job1", "success", now.Add(time.Duration(i)*time.Minute))
	}

	res, err := DetectRecovery(st, "job1", RecoveryOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsRecovered {
		t.Errorf("expected IsRecovered=true")
	}
	if res.FailureStreak != 3 {
		t.Errorf("expected streak 3, got %d", res.FailureStreak)
	}
	if res.RecoveryRunsNeeded < 2 {
		t.Errorf("expected at least 2 recovery runs, got %d", res.RecoveryRunsNeeded)
	}
}

func TestDetectRecovery_NotRecovered(t *testing.T) {
	st := New(t.TempDir() + "/r.json")
	now := time.Now()
	for i := 0; i < 4; i++ {
		addRecoveryRecord(t, st, "job2", "failure", now.Add(time.Duration(i)*time.Minute))
	}

	res, err := DetectRecovery(st, "job2", RecoveryOptions{MinFailures: 2, StableRuns: 2, MaxSamples: 50})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsRecovered {
		t.Errorf("expected IsRecovered=false")
	}
	if res.FailureStreak != 4 {
		t.Errorf("expected streak 4, got %d", res.FailureStreak)
	}
}

func TestDetectRecovery_InsufficientSamples(t *testing.T) {
	st := New(t.TempDir() + "/r.json")
	addRecoveryRecord(t, st, "job3", "failure", time.Now())

	_, err := DetectRecovery(st, "job3", RecoveryOptions{})
	if err != ErrInsufficientRecoveryData {
		t.Errorf("expected ErrInsufficientRecoveryData, got %v", err)
	}
}

func TestDetectRecovery_MissingJob(t *testing.T) {
	st := New(t.TempDir() + "/r.json")
	_, err := DetectRecovery(st, "ghost", RecoveryOptions{})
	if err == nil {
		t.Error("expected error for missing job")
	}
}

func TestDetectRecovery_NilStore(t *testing.T) {
	_, err := DetectRecovery(nil, "job", RecoveryOptions{})
	if err == nil {
		t.Error("expected error for nil store")
	}
}
