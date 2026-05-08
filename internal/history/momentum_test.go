package history

import (
	"testing"
	"time"
)

func addMomentumRecord(t *testing.T, st Store, job string, dur time.Duration, status string, at time.Time) {
	t.Helper()
	err := st.Add(Record{
		JobName:   job,
		StartedAt: at,
		Duration:  dur,
		Status:    status,
	})
	if err != nil {
		t.Fatalf("add record: %v", err)
	}
}

func TestComputeMomentum_Accelerating(t *testing.T) {
	st := New(tempPath(t))
	base := time.Now().Add(-10 * time.Minute)
	// Durations steadily increase: 100, 200, 300, 400, 500 ms
	for i := 0; i < 5; i++ {
		dur := time.Duration((i+1)*100) * time.Millisecond
		addMomentumRecord(t, st, "job", dur, "success", base.Add(time.Duration(i)*time.Minute))
	}

	res, err := ComputeMomentum(st, "job", MomentumOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Momentum <= 0 {
		t.Errorf("expected positive momentum for accelerating job, got %.2f", res.Momentum)
	}
	if res.SampleCount != 5 {
		t.Errorf("expected 5 samples, got %d", res.SampleCount)
	}
}

func TestComputeMomentum_Decelerating(t *testing.T) {
	st := New(tempPath(t))
	base := time.Now().Add(-10 * time.Minute)
	// Durations steadily decrease: 500, 400, 300, 200, 100 ms
	for i := 0; i < 5; i++ {
		dur := time.Duration((5-i)*100) * time.Millisecond
		addMomentumRecord(t, st, "job", dur, "success", base.Add(time.Duration(i)*time.Minute))
	}

	res, err := ComputeMomentum(st, "job", MomentumOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Momentum >= 0 {
		t.Errorf("expected negative momentum for decelerating job, got %.2f", res.Momentum)
	}
}

func TestComputeMomentum_InsufficientSamples(t *testing.T) {
	st := New(tempPath(t))
	base := time.Now()
	addMomentumRecord(t, st, "job", 100*time.Millisecond, "success", base)
	addMomentumRecord(t, st, "job", 200*time.Millisecond, "success", base.Add(time.Minute))

	_, err := ComputeMomentum(st, "job", MomentumOptions{})
	if err != ErrInsufficientMomentumSamples {
		t.Errorf("expected ErrInsufficientMomentumSamples, got %v", err)
	}
}

func TestComputeMomentum_IgnoresFailures(t *testing.T) {
	st := New(tempPath(t))
	base := time.Now().Add(-10 * time.Minute)
	for i := 0; i < 4; i++ {
		addMomentumRecord(t, st, "job", 200*time.Millisecond, "success", base.Add(time.Duration(i)*time.Minute))
	}
	// Add a failed run with an outlier duration — should be ignored.
	addMomentumRecord(t, st, "job", 9999*time.Millisecond, "failure", base.Add(5*time.Minute))
	addMomentumRecord(t, st, "job", 210*time.Millisecond, "success", base.Add(6*time.Minute))

	res, err := ComputeMomentum(st, "job", MomentumOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Momentum should be near-zero (stable), not skewed by the failure.
	if res.Momentum > 50 {
		t.Errorf("momentum unexpectedly high (%.2f); failure run may have been included", res.Momentum)
	}
}

func TestComputeMomentum_MissingJob(t *testing.T) {
	st := New(tempPath(t))
	_, err := ComputeMomentum(st, "nonexistent", MomentumOptions{})
	if err == nil {
		t.Error("expected error for missing job, got nil")
	}
}

func TestComputeMomentum_NilStore(t *testing.T) {
	_, err := ComputeMomentum(nil, "job", MomentumOptions{})
	if err == nil {
		t.Error("expected error for nil store")
	}
}
