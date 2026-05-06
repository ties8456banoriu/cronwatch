package history

import (
	"testing"
	"time"
)

func addTrendRecord(t *testing.T, s *Store, job string, dur time.Duration) {
	t.Helper()
	err := s.Add(Record{
		JobName:   job,
		StartTime: time.Now(),
		Duration:  dur,
		Success:   true,
	})
	if err != nil {
		t.Fatalf("add record: %v", err)
	}
}

func TestAnalyzeTrend_Degrading(t *testing.T) {
	s := New(tempPath(t))
	durations := []time.Duration{1, 2, 4, 7, 11, 16}
	for _, d := range durations {
		addTrendRecord(t, s, "backup", d*time.Second)
	}

	res, err := AnalyzeTrend(s, "backup", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Direction != TrendDegrading {
		t.Errorf("expected Degrading, got %s (slope=%.3f)", res.Direction, res.Slope)
	}
	if res.Samples != 6 {
		t.Errorf("expected 6 samples, got %d", res.Samples)
	}
}

func TestAnalyzeTrend_Improving(t *testing.T) {
	s := New(tempPath(t))
	durations := []time.Duration{20, 15, 10, 6, 3, 1}
	for _, d := range durations {
		addTrendRecord(t, s, "sync", d*time.Second)
	}

	res, err := AnalyzeTrend(s, "sync", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Direction != TrendImproving {
		t.Errorf("expected Improving, got %s (slope=%.3f)", res.Direction, res.Slope)
	}
}

func TestAnalyzeTrend_Stable(t *testing.T) {
	s := New(tempPath(t))
	for i := 0; i < 5; i++ {
		addTrendRecord(t, s, "ping", 5*time.Second)
	}

	res, err := AnalyzeTrend(s, "ping", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Direction != TrendStable {
		t.Errorf("expected Stable, got %s", res.Direction)
	}
}

func TestAnalyzeTrend_MaxSamplesRespected(t *testing.T) {
	s := New(tempPath(t))
	// First 4 records are very slow; last 3 are fast — with maxSamples=3 trend is stable/improving.
	for i := 0; i < 4; i++ {
		addTrendRecord(t, s, "job", 30*time.Second)
	}
	for i := 0; i < 3; i++ {
		addTrendRecord(t, s, "job", 5*time.Second)
	}

	res, err := AnalyzeTrend(s, "job", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Samples != 3 {
		t.Errorf("expected 3 samples, got %d", res.Samples)
	}
}

func TestAnalyzeTrend_MissingJob(t *testing.T) {
	s := New(tempPath(t))
	_, err := AnalyzeTrend(s, "ghost", 5)
	if err == nil {
		t.Error("expected error for missing job")
	}
}

func TestAnalyzeTrend_TooFewSamples(t *testing.T) {
	s := New(tempPath(t))
	_, err := AnalyzeTrend(s, "job", 1)
	if err == nil {
		t.Error("expected error when maxSamples < 2")
	}
}

func TestAnalyzeAllTrends_MultipleJobs(t *testing.T) {
	s := New(tempPath(t))
	for i := 1; i <= 4; i++ {
		addTrendRecord(t, s, "jobA", time.Duration(i)*time.Second)
		addTrendRecord(t, s, "jobB", 5*time.Second)
	}

	results, err := AnalyzeAllTrends(s, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}
