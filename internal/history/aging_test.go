package history

import (
	"testing"
	"time"
)

func addAgingRecord(t *testing.T, s *Store, job string, dur time.Duration, status string, at time.Time) {
	t.Helper()
	err := s.Add(Record{
		JobName:   job,
		StartedAt: at,
		Duration:  dur,
		Status:    status,
	})
	if err != nil {
		t.Fatalf("addAgingRecord: %v", err)
	}
}

func TestDetectAging_IsAging(t *testing.T) {
	s := New(t.TempDir() + "/aging.db")
	now := time.Now()

	// Early runs: fast (~100 ms)
	for i := 0; i < 10; i++ {
		addAgingRecord(t, s, "job", 100*time.Millisecond, "success", now.Add(-time.Duration(30-i)*24*time.Hour))
	}
	// Recent runs: slow (~200 ms) — 100 % increase
	for i := 0; i < 10; i++ {
		addAgingRecord(t, s, "job", 200*time.Millisecond, "success", now.Add(-time.Duration(10-i)*24*time.Hour))
	}

	res, err := DetectAging(s, "job", AgingOptions{WindowSize: 10, ThresholdPct: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsAging {
		t.Errorf("expected IsAging=true, got false (deltaPct=%.1f)", res.DeltaPct)
	}
	if res.DeltaPct < 90 || res.DeltaPct > 110 {
		t.Errorf("expected ~100%% delta, got %.1f", res.DeltaPct)
	}
}

func TestDetectAging_NotAging(t *testing.T) {
	s := New(t.TempDir() + "/aging2.db")
	now := time.Now()

	for i := 0; i < 20; i++ {
		addAgingRecord(t, s, "stable", 150*time.Millisecond, "success", now.Add(-time.Duration(20-i)*time.Hour))
	}

	res, err := DetectAging(s, "stable", AgingOptions{WindowSize: 10, ThresholdPct: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsAging {
		t.Errorf("expected IsAging=false for stable job")
	}
}

func TestDetectAging_InsufficientSamples(t *testing.T) {
	s := New(t.TempDir() + "/aging3.db")
	now := time.Now()

	for i := 0; i < 5; i++ {
		addAgingRecord(t, s, "sparse", 100*time.Millisecond, "success", now.Add(-time.Duration(i)*time.Hour))
	}

	_, err := DetectAging(s, "sparse", AgingOptions{WindowSize: 10})
	if err == nil {
		t.Fatal("expected error for insufficient samples")
	}
}

func TestDetectAging_IgnoresFailures(t *testing.T) {
	s := New(t.TempDir() + "/aging4.db")
	now := time.Now()

	for i := 0; i < 10; i++ {
		addAgingRecord(t, s, "mixed", 100*time.Millisecond, "success", now.Add(-time.Duration(40-i)*time.Hour))
	}
	// Interleaved failures should be ignored
	for i := 0; i < 5; i++ {
		addAgingRecord(t, s, "mixed", 9999*time.Millisecond, "failure", now.Add(-time.Duration(20-i)*time.Hour))
	}
	for i := 0; i < 10; i++ {
		addAgingRecord(t, s, "mixed", 110*time.Millisecond, "success", now.Add(-time.Duration(10-i)*time.Hour))
	}

	res, err := DetectAging(s, "mixed", AgingOptions{WindowSize: 10, ThresholdPct: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsAging {
		t.Errorf("expected IsAging=false when failures are excluded, deltaPct=%.1f", res.DeltaPct)
	}
}

func TestDetectAging_NilStore(t *testing.T) {
	_, err := DetectAging(nil, "job", AgingOptions{})
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}
