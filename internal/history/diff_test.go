package history

import (
	"testing"
	"time"
)

func addDiffRecord(t *testing.T, s *Store, name string, success bool, duration time.Duration, at time.Time) Record {
	t.Helper()
	r := Record{
		JobName:   name,
		StartTime: at,
		Duration:  duration,
		Success:   success,
	}
	if err := s.Add(r); err != nil {
		t.Fatalf("addDiffRecord: %v", err)
	}
	return r
}

func TestDiff_CompareLatest_DurationChange(t *testing.T) {
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	addDiffRecord(t, s, "backup", true, 10*time.Second, now.Add(-2*time.Minute))
	addDiffRecord(t, s, "backup", true, 25*time.Second, now.Add(-1*time.Minute))

	result, err := Diff(s, "backup", DiffOptions{CompareLatest: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DurationDiff != 15*time.Second {
		t.Errorf("expected DurationDiff=15s, got %v", result.DurationDiff)
	}
	if result.StatusChanged {
		t.Error("expected StatusChanged=false")
	}
	if result.Note != "" {
		t.Errorf("expected empty note, got %q", result.Note)
	}
}

func TestDiff_StatusDegraded(t *testing.T) {
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	addDiffRecord(t, s, "sync", true, 5*time.Second, now.Add(-2*time.Minute))
	addDiffRecord(t, s, "sync", false, 5*time.Second, now.Add(-1*time.Minute))

	result, err := Diff(s, "sync", DiffOptions{CompareLatest: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.StatusChanged {
		t.Error("expected StatusChanged=true")
	}
	if result.Note != "job degraded (was passing)" {
		t.Errorf("unexpected note: %q", result.Note)
	}
}

func TestDiff_StatusRecovered(t *testing.T) {
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	addDiffRecord(t, s, "deploy", false, 3*time.Second, now.Add(-2*time.Minute))
	addDiffRecord(t, s, "deploy", true, 3*time.Second, now.Add(-1*time.Minute))

	result, err := Diff(s, "deploy", DiffOptions{CompareLatest: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Note != "job recovered (was failing)" {
		t.Errorf("unexpected note: %q", result.Note)
	}
}

func TestDiff_ByID(t *testing.T) {
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	r1 := addDiffRecord(t, s, "etl", true, 8*time.Second, now.Add(-3*time.Minute))
	addDiffRecord(t, s, "etl", true, 12*time.Second, now.Add(-2*time.Minute))
	r3 := addDiffRecord(t, s, "etl", true, 20*time.Second, now.Add(-1*time.Minute))

	id1 := buildRecordID(r1)
	id3 := buildRecordID(r3)

	result, err := Diff(s, "etl", DiffOptions{BaselineID: id1, ComparedID: id3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DurationDiff != 12*time.Second {
		t.Errorf("expected 12s diff, got %v", result.DurationDiff)
	}
}

func TestDiff_InsufficientRecords(t *testing.T) {
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatal(err)
	}
	addDiffRecord(t, s, "lonely", true, 1*time.Second, time.Now())

	_, err = Diff(s, "lonely", DiffOptions{CompareLatest: true})
	if err == nil {
		t.Error("expected error for insufficient records")
	}
}

func TestDiff_EmptyJobName(t *testing.T) {
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatal(err)
	}
	_, err = Diff(s, "", DiffOptions{CompareLatest: true})
	if err == nil {
		t.Error("expected error for empty job name")
	}
}
