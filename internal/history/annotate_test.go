package history

import (
	"testing"
	"time"
)

func addAnnotationRecord(t *testing.T, s *Store, jobName string) {
	t.Helper()
	err := s.Add(Record{
		JobName:   jobName,
		StartedAt: time.Now().UTC(),
		Duration:  2 * time.Second,
		Success:   true,
	})
	if err != nil {
		t.Fatalf("failed to add record: %v", err)
	}
}

func TestAnnotate_Success(t *testing.T) {
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	addAnnotationRecord(t, s, "backup")

	a, err := Annotate(s, "backup", "ran after maintenance", "alice")
	if err != nil {
		t.Fatalf("Annotate: %v", err)
	}
	if a.JobName != "backup" {
		t.Errorf("expected job name %q, got %q", "backup", a.JobName)
	}
	if a.Note != "ran after maintenance" {
		t.Errorf("unexpected note: %q", a.Note)
	}
	if a.Author != "alice" {
		t.Errorf("unexpected author: %q", a.Author)
	}
	if a.RecordID == "" {
		t.Error("expected non-empty RecordID")
	}
	if a.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestAnnotate_MissingJob(t *testing.T) {
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = Annotate(s, "nonexistent", "some note", "bob")
	if err == nil {
		t.Fatal("expected error for missing job, got nil")
	}
}

func TestAnnotate_EmptyNote(t *testing.T) {
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	addAnnotationRecord(t, s, "sync")

	_, err = Annotate(s, "sync", "", "carol")
	if err == nil {
		t.Fatal("expected error for empty note, got nil")
	}
}

func TestAnnotateByID_Success(t *testing.T) {
	a, err := AnnotateByID("backup@123456", "backup", "manual trigger", "dave")
	if err != nil {
		t.Fatalf("AnnotateByID: %v", err)
	}
	if a.RecordID != "backup@123456" {
		t.Errorf("unexpected RecordID: %q", a.RecordID)
	}
}

func TestAnnotateByID_EmptyID(t *testing.T) {
	_, err := AnnotateByID("", "backup", "note", "eve")
	if err == nil {
		t.Fatal("expected error for empty record ID")
	}
}
