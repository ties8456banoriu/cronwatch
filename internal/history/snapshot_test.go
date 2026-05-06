package history

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func addSnapshotRecord(t *testing.T, s *Store, job string, dur time.Duration, success bool) {
	t.Helper()
	err := s.Add(Record{
		JobName:   job,
		StartedAt: time.Now().UTC(),
		Duration:  dur,
		Success:   success,
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
}

func TestTakeSnapshot_CapturesAllJobs(t *testing.T) {
	s := New(t.TempDir() + "/snap.db")
	addSnapshotRecord(t, s, "job-a", time.Second, true)
	addSnapshotRecord(t, s, "job-b", 2*time.Second, false)

	snap, err := TakeSnapshot(s)
	if err != nil {
		t.Fatalf("TakeSnapshot: %v", err)
	}

	if snap.JobCount != 2 {
		t.Errorf("expected 2 jobs, got %d", snap.JobCount)
	}
	if len(snap.Records["job-a"]) != 1 {
		t.Errorf("expected 1 record for job-a")
	}
	if len(snap.Records["job-b"]) != 1 {
		t.Errorf("expected 1 record for job-b")
	}
}

func TestTakeSnapshot_IsIndependent(t *testing.T) {
	s := New(t.TempDir() + "/snap.db")
	addSnapshotRecord(t, s, "job-x", time.Second, true)

	snap, _ := TakeSnapshot(s)

	// Add more records after snapshot — snapshot should not change.
	addSnapshotRecord(t, s, "job-x", 2*time.Second, true)

	if len(snap.Records["job-x"]) != 1 {
		t.Errorf("snapshot was mutated after TakeSnapshot")
	}
}

func TestSaveAndLoadSnapshot_RoundTrip(t *testing.T) {
	s := New(t.TempDir() + "/snap.db")
	addSnapshotRecord(t, s, "backup", 5*time.Second, true)

	snap, err := TakeSnapshot(s)
	if err != nil {
		t.Fatalf("TakeSnapshot: %v", err)
	}

	dir := t.TempDir()
	path, err := SaveSnapshot(snap, dir)
	if err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("snapshot file not found: %v", err)
	}

	loaded, err := LoadSnapshot(path)
	if err != nil {
		t.Fatalf("LoadSnapshot: %v", err)
	}

	if loaded.JobCount != snap.JobCount {
		t.Errorf("job count mismatch: want %d got %d", snap.JobCount, loaded.JobCount)
	}
	if len(loaded.Records["backup"]) != 1 {
		t.Errorf("expected 1 record for 'backup' after load")
	}
}

func TestSaveSnapshot_CreatesDirectory(t *testing.T) {
	s := New(t.TempDir() + "/snap.db")
	addSnapshotRecord(t, s, "job", time.Second, true)
	snap, _ := TakeSnapshot(s)

	dir := filepath.Join(t.TempDir(), "nested", "snapshots")
	_, err := SaveSnapshot(snap, dir)
	if err != nil {
		t.Fatalf("SaveSnapshot with nested dir: %v", err)
	}
}
