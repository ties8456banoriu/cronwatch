package history

import (
	"testing"
	"time"
)

func addBucketRecord(t *testing.T, s *Store, job string, dur time.Duration, status string) {
	t.Helper()
	err := s.Add(Record{
		JobName:   job,
		StartedAt: time.Now(),
		Duration:  dur,
		Status:    status,
	})
	if err != nil {
		t.Fatalf("addBucketRecord: %v", err)
	}
}

func TestBucketDurations_BasicBuckets(t *testing.T) {
	s := New(tempPath(t))
	addBucketRecord(t, s, "job1", 3*time.Second, "success")
	addBucketRecord(t, s, "job1", 7*time.Second, "success")
	addBucketRecord(t, s, "job1", 12*time.Second, "success")

	res, err := BucketDurations(s, BucketOptions{
		JobName:    "job1",
		BucketSize: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.JobName != "job1" {
		t.Errorf("expected job1, got %q", res.JobName)
	}
	if len(res.Buckets) != 3 {
		t.Errorf("expected 3 buckets, got %d", len(res.Buckets))
	}
}

func TestBucketDurations_IgnoresFailures(t *testing.T) {
	s := New(tempPath(t))
	addBucketRecord(t, s, "job2", 4*time.Second, "success")
	addBucketRecord(t, s, "job2", 4*time.Second, "failure")
	addBucketRecord(t, s, "job2", 4*time.Second, "success")

	res, err := BucketDurations(s, BucketOptions{
		JobName:    "job2",
		BucketSize: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Buckets) != 1 {
		t.Errorf("expected 1 bucket, got %d", len(res.Buckets))
	}
	if res.Buckets[0].Count != 2 {
		t.Errorf("expected count 2, got %d", res.Buckets[0].Count)
	}
}

func TestBucketDurations_MissingJob(t *testing.T) {
	s := New(tempPath(t))
	_, err := BucketDurations(s, BucketOptions{
		JobName:    "ghost",
		BucketSize: 5 * time.Second,
	})
	if err == nil {
		t.Fatal("expected error for missing job")
	}
}

func TestBucketDurations_InvalidOptions(t *testing.T) {
	s := New(tempPath(t))
	_, err := BucketDurations(s, BucketOptions{JobName: "job", BucketSize: 0})
	if err == nil {
		t.Fatal("expected error for zero bucket size")
	}
	_, err = BucketDurations(s, BucketOptions{BucketSize: 5 * time.Second})
	if err == nil {
		t.Fatal("expected error for empty job name")
	}
}

func TestBucketDurations_BucketsAreSorted(t *testing.T) {
	s := New(tempPath(t))
	addBucketRecord(t, s, "job3", 15*time.Second, "success")
	addBucketRecord(t, s, "job3", 2*time.Second, "success")
	addBucketRecord(t, s, "job3", 8*time.Second, "success")

	res, err := BucketDurations(s, BucketOptions{
		JobName:    "job3",
		BucketSize: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 1; i < len(res.Buckets); i++ {
		if res.Buckets[i].RangeStart < res.Buckets[i-1].RangeStart {
			t.Errorf("buckets not sorted at index %d", i)
		}
	}
}
