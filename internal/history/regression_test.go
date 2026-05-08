package history

import (
	"testing"
	"time"
)

func addRegressionRecord(t *testing.T, st *Store, job, status string, ms int) {
	t.Helper()
	err := st.Add(Record{
		JobName:   job,
		StartedAt: time.Now(),
		Duration:  time.Duration(ms) * time.Millisecond,
		Status:    status,
	})
	if err != nil {
		t.Fatalf("addRegressionRecord: %v", err)
	}
}

func TestComputeRegression_GrowingDurations(t *testing.T) {
	st := New(tempPath(t))
	// Durations grow linearly: 100, 110, 120, 130, 140 ms
	for i := 0; i < 5; i++ {
		addRegressionRecord(t, st, "backup", "success", 100+i*10)
	}
	res, err := ComputeRegression(st, "backup", RegressionOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Slope < 9 || res.Slope > 11 {
		t.Errorf("expected slope ~10, got %.4f", res.Slope)
	}
	if res.R2 < 0.99 {
		t.Errorf("expected R² near 1.0, got %.4f", res.R2)
	}
	if res.Warning == "" {
		t.Error("expected a warning for growing durations")
	}
}

func TestComputeRegression_StableDurations(t *testing.T) {
	st := New(tempPath(t))
	for i := 0; i < 5; i++ {
		addRegressionRecord(t, st, "sync", "success", 200)
	}
	res, err := ComputeRegression(st, "sync", RegressionOptions{SlopeWarnThreshold: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Slope > 0.1 {
		t.Errorf("expected near-zero slope, got %.4f", res.Slope)
	}
	if res.Warning != "" {
		t.Errorf("unexpected warning: %s", res.Warning)
	}
}

func TestComputeRegression_IgnoresFailures(t *testing.T) {
	st := New(tempPath(t))
	for i := 0; i < 4; i++ {
		addRegressionRecord(t, st, "etl", "success", 300)
	}
	// failed runs should be excluded
	addRegressionRecord(t, st, "etl", "failure", 9999)
	res, err := ComputeRegression(st, "etl", RegressionOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Samples != 4 {
		t.Errorf("expected 4 samples, got %d", res.Samples)
	}
}

func TestComputeRegression_InsufficientSamples(t *testing.T) {
	st := New(tempPath(t))
	addRegressionRecord(t, st, "tiny", "success", 100)
	addRegressionRecord(t, st, "tiny", "success", 110)
	_, err := ComputeRegression(st, "tiny", RegressionOptions{})
	if err == nil {
		t.Error("expected error for insufficient samples")
	}
}

func TestComputeRegression_MissingJob(t *testing.T) {
	st := New(tempPath(t))
	_, err := ComputeRegression(st, "ghost", RegressionOptions{})
	if err == nil {
		t.Error("expected error for missing job")
	}
}

func TestComputeRegression_NilStore(t *testing.T) {
	_, err := ComputeRegression(nil, "any", RegressionOptions{})
	if err == nil {
		t.Error("expected error for nil store")
	}
}

func TestComputeRegression_MaxSamples(t *testing.T) {
	st := New(tempPath(t))
	// First 3 are slow, last 3 are stable — MaxSamples=3 should pick stable ones
	for i := 0; i < 3; i++ {
		addRegressionRecord(t, st, "job", "success", 500+i*100)
	}
	for i := 0; i < 3; i++ {
		addRegressionRecord(t, st, "job", "success", 200)
	}
	res, err := ComputeRegression(st, "job", RegressionOptions{MaxSamples: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Samples != 3 {
		t.Errorf("expected 3 samples, got %d", res.Samples)
	}
	if res.Slope > 1 {
		t.Errorf("expected near-zero slope for stable window, got %.4f", res.Slope)
	}
}
