package history

import (
	"testing"
	"time"
)

func addClusterRecord(t *testing.T, s *Store, job string, dur time.Duration, success bool) {
	t.Helper()
	err := s.Add(Record{
		JobName:  job,
		Start:    time.Now(),
		Duration: dur,
		Success:  success,
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
}

func TestCluster_TwoGroups(t *testing.T) {
	s := New(tempPath(t))

	// Two clear clusters: fast (~1s) and slow (~10s).
	for i := 0; i < 5; i++ {
		addClusterRecord(t, s, "backup", time.Second+time.Duration(i)*10*time.Millisecond, true)
		addClusterRecord(t, s, "backup", 10*time.Second+time.Duration(i)*10*time.Millisecond, true)
	}

	res, err := Cluster(s, "backup", ClusterOptions{K: 2})
	if err != nil {
		t.Fatalf("Cluster: %v", err)
	}
	if res.JobName != "backup" {
		t.Errorf("JobName = %q, want backup", res.JobName)
	}
	if len(res.Centroids) != 2 {
		t.Fatalf("len(Centroids) = %d, want 2", len(res.Centroids))
	}
	// Centroids should be approximately 1s and 10s.
	low, high := res.Centroids[0], res.Centroids[1]
	if low > high {
		low, high = high, low
	}
	if low > 2*time.Second {
		t.Errorf("low centroid = %v, want ~1s", low)
	}
	if high < 8*time.Second {
		t.Errorf("high centroid = %v, want ~10s", high)
	}
	if res.Inertia < 0 {
		t.Errorf("Inertia should be non-negative")
	}
}

func TestCluster_InsufficientSamples(t *testing.T) {
	s := New(tempPath(t))
	addClusterRecord(t, s, "job", time.Second, true)

	_, err := Cluster(s, "job", ClusterOptions{K: 3})
	if err != ErrInsufficientData {
		t.Errorf("err = %v, want ErrInsufficientData", err)
	}
}

func TestCluster_MissingJob(t *testing.T) {
	s := New(tempPath(t))

	_, err := Cluster(s, "ghost", ClusterOptions{K: 2})
	if err != ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestCluster_IgnoresFailedRuns(t *testing.T) {
	s := New(tempPath(t))

	// 4 successful + 10 failed; only successes should be clustered.
	for i := 0; i < 4; i++ {
		addClusterRecord(t, s, "job", time.Second, true)
	}
	for i := 0; i < 10; i++ {
		addClusterRecord(t, s, "job", 30*time.Second, false)
	}

	res, err := Cluster(s, "job", ClusterOptions{K: 2})
	if err != nil {
		t.Fatalf("Cluster: %v", err)
	}
	if len(res.Labels) != 4 {
		t.Errorf("Labels len = %d, want 4 (only successful runs)", len(res.Labels))
	}
}

func TestCluster_InvalidK(t *testing.T) {
	s := New(tempPath(t))
	addClusterRecord(t, s, "job", time.Second, true)

	_, err := Cluster(s, "job", ClusterOptions{K: 1})
	if err == nil {
		t.Error("expected error for K < 2")
	}
}
