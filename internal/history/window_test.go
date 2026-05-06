package history

import (
	"testing"
	"time"
)

func addWindowRecord(t *testing.T, s *Store, job string, ago time.Duration, dur time.Duration, success bool) {
	t.Helper()
	r := Record{
		JobName:   job,
		StartedAt: time.Now().UTC().Add(-ago),
		Duration:  dur,
		Success:   success,
	}
	if err := s.Add(r); err != nil {
		t.Fatalf("addWindowRecord: %v", err)
	}
}

func TestSlidingWindow_BasicBuckets(t *testing.T) {
	s := New(tempPath(t))
	addWindowRecord(t, s, "job1", 30*time.Minute, 2*time.Second, true)
	addWindowRecord(t, s, "job1", 90*time.Minute, 4*time.Second, true)
	addWindowRecord(t, s, "job1", 150*time.Minute, 3*time.Second, false)

	buckets, err := SlidingWindow(s, WindowOptions{
		JobName:  "job1",
		Duration: 4 * time.Hour,
		Step:     time.Hour,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(buckets) != 4 {
		t.Fatalf("expected 4 buckets, got %d", len(buckets))
	}
	total := 0
	for _, b := range buckets {
		total += b.Count
	}
	if total != 3 {
		t.Errorf("expected 3 total records across buckets, got %d", total)
	}
}

func TestSlidingWindow_FailuresTracked(t *testing.T) {
	s := New(tempPath(t))
	addWindowRecord(t, s, "job2", 10*time.Minute, time.Second, false)
	addWindowRecord(t, s, "job2", 20*time.Minute, time.Second, false)
	addWindowRecord(t, s, "job2", 30*time.Minute, time.Second, true)

	buckets, err := SlidingWindow(s, WindowOptions{
		JobName:  "job2",
		Duration: time.Hour,
		Step:     time.Hour,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(buckets))
	}
	if buckets[0].Failures != 2 {
		t.Errorf("expected 2 failures, got %d", buckets[0].Failures)
	}
}

func TestSlidingWindow_MissingJob(t *testing.T) {
	s := New(tempPath(t))
	buckets, err := SlidingWindow(s, WindowOptions{
		JobName:  "ghost",
		Duration: time.Hour,
		Step:     time.Minute,
	})
	if err != nil {
		t.Fatalf("unexpected error for missing job: %v", err)
	}
	for _, b := range buckets {
		if b.Count != 0 {
			t.Errorf("expected empty buckets for missing job")
		}
	}
}

func TestSlidingWindow_InvalidOptions(t *testing.T) {
	s := New(tempPath(t))
	_, err := SlidingWindow(s, WindowOptions{JobName: "", Duration: time.Hour, Step: time.Minute})
	if err == nil {
		t.Error("expected error for empty job name")
	}
	_, err = SlidingWindow(s, WindowOptions{JobName: "j", Duration: 0, Step: time.Minute})
	if err == nil {
		t.Error("expected error for zero duration")
	}
	_, err = SlidingWindow(s, WindowOptions{JobName: "j", Duration: time.Hour, Step: 0})
	if err == nil {
		t.Error("expected error for zero step")
	}
}
