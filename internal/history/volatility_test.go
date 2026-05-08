package history

import (
	"testing"
	"time"
)

func addVolatilityRecord(t *testing.T, st *Store, job string, ms int64, ok bool) {
	t.Helper()
	status := "success"
	if !ok {
		status = "failure"
	}
	err := st.Add(Record{
		JobName:   job,
		StartedAt: time.Now(),
		Duration:  time.Duration(ms) * time.Millisecond,
		Status:    status,
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
}

func TestComputeVolatility_HighCV(t *testing.T) {
	st := New(tempPath(t))
	// large spread → high CV
	for _, ms := range []int64{100, 900, 100, 900, 100} {
		addVolatilityRecord(t, st, "job", ms, true)
	}
	res, err := ComputeVolatility(st, "job", VolatilityOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsVolatile {
		t.Errorf("expected job to be volatile, CV=%.3f", res.CV)
	}
	if res.SampleSize != 5 {
		t.Errorf("expected 5 samples, got %d", res.SampleSize)
	}
}

func TestComputeVolatility_LowCV(t *testing.T) {
	st := New(tempPath(t))
	// tight cluster → low CV
	for _, ms := range []int64{200, 201, 199, 200, 202} {
		addVolatilityRecord(t, st, "job", ms, true)
	}
	res, err := ComputeVolatility(st, "job", VolatilityOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsVolatile {
		t.Errorf("expected job to be stable, CV=%.3f", res.CV)
	}
}

func TestComputeVolatility_IgnoresFailures(t *testing.T) {
	st := New(tempPath(t))
	// failures should not skew the calculation
	for _, ms := range []int64{200, 201, 200} {
		addVolatilityRecord(t, st, "job", ms, true)
	}
	addVolatilityRecord(t, st, "job", 9999, false)
	res, err := ComputeVolatility(st, "job", VolatilityOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsVolatile {
		t.Errorf("failure record should be ignored; CV=%.3f", res.CV)
	}
}

func TestComputeVolatility_InsufficientSamples(t *testing.T) {
	st := New(tempPath(t))
	addVolatilityRecord(t, st, "job", 300, true)
	_, err := ComputeVolatility(st, "job", VolatilityOptions{})
	if err == nil {
		t.Fatal("expected error for insufficient samples")
	}
}

func TestComputeVolatility_MaxSamples(t *testing.T) {
	st := New(tempPath(t))
	// first two are wildly different, last three are stable
	for _, ms := range []int64{1, 9999, 200, 201, 200} {
		addVolatilityRecord(t, st, "job", ms, true)
	}
	res, err := ComputeVolatility(st, "job", VolatilityOptions{MaxSamples: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.SampleSize != 3 {
		t.Errorf("expected 3 samples, got %d", res.SampleSize)
	}
	if res.IsVolatile {
		t.Errorf("expected stable with MaxSamples=3, CV=%.3f", res.CV)
	}
}

func TestComputeVolatility_NilStore(t *testing.T) {
	_, err := ComputeVolatility(nil, "job", VolatilityOptions{})
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}
