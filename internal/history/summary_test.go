package history

import (
	"testing"
	"time"
)

func makeRecords() []Record {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	return []Record{
		{JobName: "job", StartedAt: base, Duration: 2 * time.Second, Success: true},
		{JobName: "job", StartedAt: base.Add(time.Hour), Duration: 4 * time.Second, Success: true},
		{JobName: "job", StartedAt: base.Add(2 * time.Hour), Duration: 6 * time.Second, Success: false},
	}
}

func TestSummarize_Counts(t *testing.T) {
	s := Summarize("job", makeRecords())
	if s.TotalRuns != 3 {
		t.Errorf("expected TotalRuns=3, got %d", s.TotalRuns)
	}
	if s.SuccessRuns != 2 {
		t.Errorf("expected SuccessRuns=2, got %d", s.SuccessRuns)
	}
	if s.FailureRuns != 1 {
		t.Errorf("expected FailureRuns=1, got %d", s.FailureRuns)
	}
}

func TestSummarize_Durations(t *testing.T) {
	s := Summarize("job", makeRecords())
	if s.AvgDuration != 4*time.Second {
		t.Errorf("expected AvgDuration=4s, got %v", s.AvgDuration)
	}
	if s.MaxDuration != 6*time.Second {
		t.Errorf("expected MaxDuration=6s, got %v", s.MaxDuration)
	}
}

func TestSummarize_LastRun(t *testing.T) {
	records := makeRecords()
	s := Summarize("job", records)
	expected := records[2].StartedAt
	if !s.LastRun.Equal(expected) {
		t.Errorf("expected LastRun=%v, got %v", expected, s.LastRun)
	}
}

func TestSummarize_Empty(t *testing.T) {
	s := Summarize("job", nil)
	if s.TotalRuns != 0 || s.AvgDuration != 0 {
		t.Errorf("expected zero summary for empty records, got %+v", s)
	}
}
