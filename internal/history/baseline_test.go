package history_test

import (
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/history"
)

func addBaselineRecord(t *testing.T, s *history.Store, job string, dur time.Duration, success bool) {
	t.Helper()
	err := s.Add(history.Record{
		JobName:   job,
		StartedAt: time.Now().UTC(),
		Duration:  dur,
		Success:   success,
	})
	if err != nil {
		t.Fatalf("addBaselineRecord: %v", err)
	}
}

func TestCaptureBaseline_Average(t *testing.T) {
	p := tempPath(t)
	s := history.New(p)

	addBaselineRecord(t, s, "myjob", 10*time.Second, true)
	addBaselineRecord(t, s, "myjob", 20*time.Second, true)
	addBaselineRecord(t, s, "myjob", 30*time.Second, true)

	b, err := history.CaptureBaseline(s, "myjob", 3, 0.10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.ExpectedDuration != 20*time.Second {
		t.Errorf("expected 20s average, got %v", b.ExpectedDuration)
	}
	if b.SampleSize != 3 {
		t.Errorf("expected sample size 3, got %d", b.SampleSize)
	}
	if b.TolerancePct != 0.10 {
		t.Errorf("expected tolerance 0.10, got %f", b.TolerancePct)
	}
}

func TestCaptureBaseline_LimitsSamples(t *testing.T) {
	p := tempPath(t)
	s := history.New(p)

	for i := 0; i < 5; i++ {
		addBaselineRecord(t, s, "job", 10*time.Second, true)
	}

	b, err := history.CaptureBaseline(s, "job", 2, 0.20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.SampleSize != 2 {
		t.Errorf("expected sample size 2, got %d", b.SampleSize)
	}
}

func TestCaptureBaseline_MissingJob(t *testing.T) {
	p := tempPath(t)
	s := history.New(p)

	_, err := history.CaptureBaseline(s, "ghost", 5, 0.10)
	if err == nil {
		t.Fatal("expected error for missing job")
	}
}

func TestCompareToBaseline_Within(t *testing.T) {
	b := history.Baseline{
		JobName:         "job",
		ExpectedDuration: 10 * time.Second,
		TolerancePct:    0.20,
	}
	r := history.Record{JobName: "job", Duration: 11 * time.Second, Success: true}
	res := history.CompareToBaseline(r, b)
	if !res.Within {
		t.Errorf("expected within tolerance, delta=%v", res.Delta)
	}
}

func TestCompareToBaseline_Outside(t *testing.T) {
	b := history.Baseline{
		JobName:         "job",
		ExpectedDuration: 10 * time.Second,
		TolerancePct:    0.10,
	}
	r := history.Record{JobName: "job", Duration: 15 * time.Second, Success: true}
	res := history.CompareToBaseline(r, b)
	if res.Within {
		t.Errorf("expected outside tolerance, delta=%v", res.Delta)
	}
	if res.DeltaPct < 0.49 {
		t.Errorf("expected deltaPct ~0.5, got %f", res.DeltaPct)
	}
}

func TestSaveAndLoadBaseline_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	b := history.Baseline{
		JobName:         "roundtrip",
		ExpectedDuration: 5 * time.Minute,
		TolerancePct:    0.15,
		SampleSize:      10,
	}
	if err := history.SaveBaseline(dir, b); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded, err := history.LoadBaseline(dir, "roundtrip")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.ExpectedDuration != b.ExpectedDuration {
		t.Errorf("duration mismatch: got %v, want %v", loaded.ExpectedDuration, b.ExpectedDuration)
	}
	if loaded.TolerancePct != b.TolerancePct {
		t.Errorf("tolerance mismatch: got %f, want %f", loaded.TolerancePct, b.TolerancePct)
	}
}

func TestLoadBaseline_Missing(t *testing.T) {
	dir := t.TempDir()
	_, err := history.LoadBaseline(dir, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing baseline file")
	}
}
