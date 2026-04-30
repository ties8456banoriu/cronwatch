package history

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"
)

func TestExportCSV_Headers(t *testing.T) {
	path := tempPath(t)
	s := New(path)

	s.Add(Record{JobName: "backup", StartedAt: time.Now(), Duration: 2 * time.Second, Success: true})

	var buf bytes.Buffer
	if err := s.ExportCSV("backup", &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := csv.NewReader(&buf)
	header, err := r.Read()
	if err != nil {
		t.Fatalf("reading header: %v", err)
	}

	want := []string{"job_name", "started_at", "duration_ms", "success"}
	for i, h := range want {
		if header[i] != h {
			t.Errorf("header[%d] = %q, want %q", i, header[i], h)
		}
	}
}

func TestExportCSV_RowValues(t *testing.T) {
	path := tempPath(t)
	s := New(path)

	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	s.Add(Record{JobName: "sync", StartedAt: now, Duration: 500 * time.Millisecond, Success: false})

	var buf bytes.Buffer
	if err := s.ExportCSV("sync", &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := csv.NewReader(&buf)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("reading csv: %v", err)
	}

	if len(records) != 2 { // header + 1 row
		t.Fatalf("expected 2 rows, got %d", len(records))
	}

	row := records[1]
	if row[0] != "sync" {
		t.Errorf("job_name = %q, want %q", row[0], "sync")
	}
	if row[2] != "500" {
		t.Errorf("duration_ms = %q, want \"500\"", row[2])
	}
	if row[3] != "false" {
		t.Errorf("success = %q, want \"false\"", row[3])
	}
}

func TestExportCSV_MissingJob(t *testing.T) {
	path := tempPath(t)
	s := New(path)

	var buf bytes.Buffer
	err := s.ExportCSV("nonexistent", &buf)
	if err == nil {
		t.Fatal("expected error for missing job, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error message should mention job name, got: %v", err)
	}
}

func TestExportAllCSV_MultipleJobs(t *testing.T) {
	path := tempPath(t)
	s := New(path)

	now := time.Now()
	s.Add(Record{JobName: "alpha", StartedAt: now, Duration: time.Second, Success: true})
	s.Add(Record{JobName: "beta", StartedAt: now, Duration: 2 * time.Second, Success: true})

	var buf bytes.Buffer
	if err := s.ExportAllCSV(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := csv.NewReader(&buf)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("reading csv: %v", err)
	}

	// header + 2 data rows
	if len(records) != 3 {
		t.Fatalf("expected 3 rows (1 header + 2 data), got %d", len(records))
	}
}
