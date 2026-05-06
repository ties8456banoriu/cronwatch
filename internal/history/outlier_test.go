package history

import (
	"testing"
	"time"
)

func addOutlierRecord(t *testing.T, s *Store, job string, dur time.Duration, status string) {
	t.Helper()
	err := s.Add(Record{
		JobName:   job,
		StartTime: time.Now(),
		Duration:  dur,
		Status:    status,
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
}

func TestDetectOutliers_FlagsSlowRun(t *testing.T) {
	s := New(t.TempDir() + "/outlier.db")
	base := 5 * time.Second
	for i := 0; i < 8; i++ {
		addOutlierRecord(t, s, "job1", base, "success")
	}
	// One very slow run
	addOutlierRecord(t, s, "job1", 30*time.Second, "success")

	results, err := DetectOutliers(s, "job1", OutlierOptions{ZScoreThreshold: 2.0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one outlier, got none")
	}
	if !results[len(results)-1].IsHigh {
		t.Error("expected outlier to be flagged as high duration")
	}
}

func TestDetectOutliers_NoOutliersWhenUniform(t *testing.T) {
	s := New(t.TempDir() + "/outlier2.db")
	for i := 0; i < 10; i++ {
		addOutlierRecord(t, s, "job2", 5*time.Second, "success")
	}

	results, err := DetectOutliers(s, "job2", OutlierOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no outliers, got %d", len(results))
	}
}

func TestDetectOutliers_InsufficientSamples(t *testing.T) {
	s := New(t.TempDir() + "/outlier3.db")
	addOutlierRecord(t, s, "job3", 5*time.Second, "success")
	addOutlierRecord(t, s, "job3", 6*time.Second, "success")

	results, err := DetectOutliers(s, "job3", OutlierOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results with insufficient samples, got %v", results)
	}
}

func TestDetectOutliers_IgnoresFailedRuns(t *testing.T) {
	s := New(t.TempDir() + "/outlier4.db")
	for i := 0; i < 5; i++ {
		addOutlierRecord(t, s, "job4", 5*time.Second, "success")
	}
	// Failed runs with extreme duration should not affect results
	for i := 0; i < 5; i++ {
		addOutlierRecord(t, s, "job4", 120*time.Second, "failure")
	}
	addOutlierRecord(t, s, "job4", 5*time.Second, "success")

	results, err := DetectOutliers(s, "job4", OutlierOptions{ZScoreThreshold: 2.0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no outliers among success runs, got %d", len(results))
	}
}
