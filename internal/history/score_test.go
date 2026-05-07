package history

import (
	"testing"
	"time"
)

func addScoreRecord(t *testing.T, store *Store, job, status string, dur time.Duration) {
	t.Helper()
	err := store.Add(Record{
		JobName:   job,
		Status:    status,
		StartedAt: time.Now(),
		Duration:  dur,
	})
	if err != nil {
		t.Fatalf("add record: %v", err)
	}
}

func TestScoreJob_PerfectScore(t *testing.T) {
	store := New(tempPath(t))
	for i := 0; i < 10; i++ {
		addScoreRecord(t, store, "backup", "success", 5*time.Second)
	}
	result, err := ScoreJob(store, "backup", ScoreOptions{SlowThreshold: 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Score < 90 {
		t.Errorf("expected high score, got %.2f", result.Score)
	}
	if result.Grade != "A" {
		t.Errorf("expected grade A, got %s", result.Grade)
	}
}

func TestScoreJob_LowSuccessRate(t *testing.T) {
	store := New(tempPath(t))
	for i := 0; i < 8; i++ {
		addScoreRecord(t, store, "etl", "failure", 1*time.Second)
	}
	for i := 0; i < 2; i++ {
		addScoreRecord(t, store, "etl", "success", 1*time.Second)
	}
	result, err := ScoreJob(store, "etl", ScoreOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Score >= 60 {
		t.Errorf("expected low score for 20%% success rate, got %.2f", result.Score)
	}
	if result.SuccessRate > 0.25 {
		t.Errorf("expected success rate ~0.2, got %.2f", result.SuccessRate)
	}
}

func TestScoreJob_SlowRunsPenalised(t *testing.T) {
	store := New(tempPath(t))
	for i := 0; i < 10; i++ {
		addScoreRecord(t, store, "report", "success", 120*time.Second)
	}
	result, err := ScoreJob(store, "report", ScoreOptions{SlowThreshold: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Penalty == 0 {
		t.Error("expected non-zero penalty for slow runs")
	}
}

func TestScoreJob_MissingJob(t *testing.T) {
	store := New(tempPath(t))
	_, err := ScoreJob(store, "ghost", ScoreOptions{})
	if err == nil {
		t.Error("expected error for missing job")
	}
}

func TestScoreJob_MaxSamplesRespected(t *testing.T) {
	store := New(tempPath(t))
	// Add 20 failures then 5 successes
	for i := 0; i < 20; i++ {
		addScoreRecord(t, store, "sync", "failure", time.Second)
	}
	for i := 0; i < 5; i++ {
		addScoreRecord(t, store, "sync", "success", time.Second)
	}
	// With MaxSamples=5 only the recent successes are seen
	result, err := ScoreJob(store, "sync", ScoreOptions{MaxSamples: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SuccessRate != 1.0 {
		t.Errorf("expected 100%% success rate with MaxSamples=5, got %.2f", result.SuccessRate)
	}
}
