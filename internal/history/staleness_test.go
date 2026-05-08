package history

import (
	"testing"
	"time"
)

func addStalenessRecord(t *testing.T, store *Store, jobName, status string, start time.Time, dur time.Duration) {
	t.Helper()
	err := store.Add(Record{
		JobName:   jobName,
		Status:    status,
		StartedAt: start,
		Duration:  dur,
	})
	if err != nil {
		t.Fatalf("addStalenessRecord: %v", err)
	}
}

func TestDetectStaleness_NotStale(t *testing.T) {
	store := New(tempPath(t))
	now := time.Now()
	addStalenessRecord(t, store, "myjob", "success", now.Add(-1*time.Hour), 2*time.Second)

	res, err := DetectStaleness(store, "myjob", StalenessOptions{Threshold: 25 * time.Hour, Now: now})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsStale {
		t.Errorf("expected job to be fresh, got stale")
	}
	if res.Age < time.Hour-time.Second || res.Age > time.Hour+time.Second {
		t.Errorf("unexpected age: %v", res.Age)
	}
}

func TestDetectStaleness_IsStale(t *testing.T) {
	store := New(tempPath(t))
	now := time.Now()
	addStalenessRecord(t, store, "myjob", "success", now.Add(-30*time.Hour), 2*time.Second)

	res, err := DetectStaleness(store, "myjob", StalenessOptions{Threshold: 25 * time.Hour, Now: now})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsStale {
		t.Errorf("expected job to be stale")
	}
}

func TestDetectStaleness_IgnoresFailedRuns(t *testing.T) {
	store := New(tempPath(t))
	now := time.Now()
	// Only failed runs — last success was long ago
	addStalenessRecord(t, store, "myjob", "success", now.Add(-48*time.Hour), 2*time.Second)
	addStalenessRecord(t, store, "myjob", "failure", now.Add(-1*time.Hour), 1*time.Second)

	res, err := DetectStaleness(store, "myjob", StalenessOptions{Threshold: 25 * time.Hour, Now: now})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsStale {
		t.Errorf("expected stale because last success was 48h ago")
	}
}

func TestDetectStaleness_MissingJob(t *testing.T) {
	store := New(tempPath(t))
	_, err := DetectStaleness(store, "ghost", StalenessOptions{})
	if err == nil {
		t.Fatal("expected error for missing job")
	}
}

func TestDetectStaleness_DefaultThreshold(t *testing.T) {
	store := New(tempPath(t))
	now := time.Now()
	addStalenessRecord(t, store, "myjob", "success", now.Add(-26*time.Hour), 1*time.Second)

	// Threshold zero → defaults to 25h
	res, err := DetectStaleness(store, "myjob", StalenessOptions{Now: now})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsStale {
		t.Errorf("expected stale with default 25h threshold")
	}
	if res.Threshold != 25*time.Hour {
		t.Errorf("expected default threshold 25h, got %v", res.Threshold)
	}
}
