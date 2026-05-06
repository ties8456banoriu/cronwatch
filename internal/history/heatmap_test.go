package history

import (
	"testing"
	"time"
)

func addHeatmapRecord(t *testing.T, s *Store, job string, start time.Time, dur time.Duration, status string) {
	t.Helper()
	err := s.Add(Record{
		JobName:   job,
		StartedAt: start,
		Duration:  dur,
		Status:    status,
	})
	if err != nil {
		t.Fatalf("addHeatmapRecord: %v", err)
	}
}

func TestBuildHeatmap_BasicBuckets(t *testing.T) {
	s := New(tempPath(t))
	now := time.Now().Truncate(time.Hour)

	addHeatmapRecord(t, s, "job1", now, 10*time.Second, "ok")
	addHeatmapRecord(t, s, "job1", now.Add(10*time.Minute), 20*time.Second, "ok")
	addHeatmapRecord(t, s, "job1", now.Add(1*time.Hour), 30*time.Second, "failed")

	hm, err := BuildHeatmap(s, "job1", HeatmapOptions{
		Granule: time.Hour,
		Since:   now,
		Until:   now.Add(2 * time.Hour),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hm.JobName != "job1" {
		t.Errorf("expected job1, got %s", hm.JobName)
	}
	if len(hm.Cells) == 0 {
		t.Fatal("expected non-empty cells")
	}
	// First bucket should have 2 runs.
	if hm.Cells[0].Count != 2 {
		t.Errorf("expected 2 runs in first bucket, got %d", hm.Cells[0].Count)
	}
}

func TestBuildHeatmap_FailuresTracked(t *testing.T) {
	s := New(tempPath(t))
	now := time.Now().Truncate(time.Hour)

	addHeatmapRecord(t, s, "job2", now, 5*time.Second, "failed")
	addHeatmapRecord(t, s, "job2", now.Add(5*time.Minute), 5*time.Second, "failed")
	addHeatmapRecord(t, s, "job2", now.Add(10*time.Minute), 5*time.Second, "ok")

	hm, err := BuildHeatmap(s, "job2", HeatmapOptions{
		Granule: time.Hour,
		Since:   now,
		Until:   now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hm.Cells[0].Failures != 2 {
		t.Errorf("expected 2 failures, got %d", hm.Cells[0].Failures)
	}
}

func TestBuildHeatmap_AvgDuration(t *testing.T) {
	s := New(tempPath(t))
	now := time.Now().Truncate(time.Hour)

	addHeatmapRecord(t, s, "job3", now, 10*time.Second, "ok")
	addHeatmapRecord(t, s, "job3", now.Add(5*time.Minute), 30*time.Second, "ok")

	hm, err := BuildHeatmap(s, "job3", HeatmapOptions{
		Granule: time.Hour,
		Since:   now,
		Until:   now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := 20 * time.Second
	if hm.Cells[0].AvgDuration != want {
		t.Errorf("expected avg %v, got %v", want, hm.Cells[0].AvgDuration)
	}
}

func TestBuildHeatmap_MissingJob(t *testing.T) {
	s := New(tempPath(t))
	_, err := BuildHeatmap(s, "ghost", HeatmapOptions{Granule: time.Hour})
	if err == nil {
		t.Fatal("expected error for missing job")
	}
}

func TestBuildHeatmap_InvalidGranule(t *testing.T) {
	s := New(tempPath(t))
	_, err := BuildHeatmap(s, "job", HeatmapOptions{Granule: 0})
	if err == nil {
		t.Fatal("expected error for zero granule")
	}
}
