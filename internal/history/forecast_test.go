package history

import (
	"testing"
	"time"
)

func addForecastRecord(t *testing.T, s *Store, job string, durationMs int64, status string) {
	t.Helper()
	err := s.Add(Record{
		JobName:    job,
		StartedAt:  time.Now().UTC(),
		DurationMs: durationMs,
		Status:     status,
	})
	if err != nil {
		t.Fatalf("addForecastRecord: %v", err)
	}
}

func TestForecast_BasicEstimate(t *testing.T) {
	s := New(tempPath(t))
	for _, d := range []int64{100, 110, 105, 108, 102} {
		addForecastRecord(t, s, "job1", d, "success")
	}

	result, err := Forecast(s, "job1", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.JobName != "job1" {
		t.Errorf("expected job1, got %s", result.JobName)
	}
	if result.EstimatedMs < 100 || result.EstimatedMs > 115 {
		t.Errorf("estimate out of expected range: %.2f", result.EstimatedMs)
	}
	if result.LowerBoundMs > result.EstimatedMs {
		t.Errorf("lower bound should be <= estimated")
	}
	if result.UpperBoundMs < result.EstimatedMs {
		t.Errorf("upper bound should be >= estimated")
	}
	if result.SampleCount != 5 {
		t.Errorf("expected 5 samples, got %d", result.SampleCount)
	}
}

func TestForecast_InsufficientSamples(t *testing.T) {
	s := New(tempPath(t))
	addForecastRecord(t, s, "job2", 200, "success")
	addForecastRecord(t, s, "job2", 210, "success")

	_, err := Forecast(s, "job2", 10)
	if err == nil {
		t.Fatal("expected error for insufficient samples")
	}
}

func TestForecast_IgnoresFailedRuns(t *testing.T) {
	s := New(tempPath(t))
	for _, d := range []int64{100, 105, 102} {
		addForecastRecord(t, s, "job3", d, "success")
	}
	// These should be ignored.
	addForecastRecord(t, s, "job3", 9999, "failure")
	addForecastRecord(t, s, "job3", 8888, "failure")

	result, err := Forecast(s, "job3", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.EstimatedMs > 200 {
		t.Errorf("failed runs should not inflate estimate: %.2f", result.EstimatedMs)
	}
}

func TestForecast_TrendDegrading(t *testing.T) {
	s := New(tempPath(t))
	for _, d := range []int64{100, 110, 130, 160, 200, 250} {
		addForecastRecord(t, s, "job4", d, "success")
	}

	result, err := Forecast(s, "job4", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TrendDirection != "degrading" {
		t.Errorf("expected degrading trend, got %s", result.TrendDirection)
	}
}

func TestForecast_TrendImproving(t *testing.T) {
	s := New(tempPath(t))
	for _, d := range []int64{250, 200, 160, 130, 110, 100} {
		addForecastRecord(t, s, "job5", d, "success")
	}

	result, err := Forecast(s, "job5", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TrendDirection != "improving" {
		t.Errorf("expected improving trend, got %s", result.TrendDirection)
	}
}

func TestForecast_MissingJob(t *testing.T) {
	s := New(tempPath(t))
	_, err := Forecast(s, "nonexistent", 10)
	if err == nil {
		t.Fatal("expected error for missing job")
	}
}
