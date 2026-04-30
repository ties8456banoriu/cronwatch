package history

import (
	"testing"
	"time"
)

func TestDefaultRetentionPolicy(t *testing.T) {
	p := DefaultRetentionPolicy()
	if p.MaxAge != 30*24*time.Hour {
		t.Errorf("expected MaxAge 30 days, got %v", p.MaxAge)
	}
	if p.MaxRecords != 500 {
		t.Errorf("expected MaxRecords 500, got %d", p.MaxRecords)
	}
}

func TestRetentionPolicy_Validate_Valid(t *testing.T) {
	p := DefaultRetentionPolicy()
	if err := p.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRetentionPolicy_Validate_NegativeAge(t *testing.T) {
	p := RetentionPolicy{MaxAge: -1 * time.Hour, MaxRecords: 10}
	if err := p.Validate(); err == nil {
		t.Error("expected error for negative MaxAge, got nil")
	}
}

func TestRetentionPolicy_Validate_NegativeRecords(t *testing.T) {
	p := RetentionPolicy{MaxAge: time.Hour, MaxRecords: -1}
	if err := p.Validate(); err == nil {
		t.Error("expected error for negative MaxRecords, got nil")
	}
}

func TestRetentionPolicy_Apply(t *testing.T) {
	path := tempPath(t)
	s := New(path)

	now := time.Now()
	old := now.Add(-40 * 24 * time.Hour)

	// Add an old record and a recent record for job "backup"
	_ = s.Add("backup", Record{StartedAt: old, Duration: time.Second, Success: true})
	_ = s.Add("backup", Record{StartedAt: now, Duration: time.Second, Success: true})

	p := RetentionPolicy{
		MaxAge:     30 * 24 * time.Hour,
		MaxRecords: 100,
	}

	if err := p.Apply(s); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	records, err := s.All("backup")
	if err != nil {
		t.Fatalf("All returned error: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("expected 1 record after retention, got %d", len(records))
	}
	if !records[0].StartedAt.Equal(now) {
		t.Errorf("expected remaining record to be the recent one")
	}
}

func TestRetentionPolicy_Apply_MaxRecords(t *testing.T) {
	path := tempPath(t)
	s := New(path)

	now := time.Now()
	for i := 0; i < 5; i++ {
		_ = s.Add("sync", Record{
			StartedAt: now.Add(-time.Duration(i) * time.Minute),
			Duration:  time.Second,
			Success:   true,
		})
	}

	p := RetentionPolicy{MaxAge: 0, MaxRecords: 3}
	if err := p.Apply(s); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	records, err := s.All("sync")
	if err != nil {
		t.Fatalf("All returned error: %v", err)
	}
	if len(records) != 3 {
		t.Errorf("expected 3 records after retention, got %d", len(records))
	}
}
