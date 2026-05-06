package history

import (
	"testing"
	"time"
)

func setupSearchStore(t *testing.T) *Store {
	t.Helper()
	s, err := New(t.TempDir() + "/search.db")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func addSearchRecord(t *testing.T, s *Store, name string, r Record) {
	t.Helper()
	if err := s.Add(name, r); err != nil {
		t.Fatalf("Add: %v", err)
	}
}

func TestSearch_ByJobName(t *testing.T) {
	s := setupSearchStore(t)
	now := time.Now()
	addSearchRecord(t, s, "backup", Record{StartedAt: now, Duration: time.Second, Success: true})
	addSearchRecord(t, s, "cleanup", Record{StartedAt: now, Duration: time.Second, Success: true})

	results, err := Search(s, SearchQuery{JobName: "backup"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 || results[0].JobName != "backup" {
		t.Errorf("expected 1 backup result, got %+v", results)
	}
}

func TestSearch_ByStatus(t *testing.T) {
	s := setupSearchStore(t)
	now := time.Now()
	addSearchRecord(t, s, "job", Record{StartedAt: now, Success: true})
	addSearchRecord(t, s, "job", Record{StartedAt: now.Add(time.Minute), Success: false})

	results, err := Search(s, SearchQuery{Status: "failure"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 || results[0].Record.Success {
		t.Errorf("expected 1 failed result, got %+v", results)
	}
}

func TestSearch_WithLimit(t *testing.T) {
	s := setupSearchStore(t)
	now := time.Now()
	for i := 0; i < 5; i++ {
		addSearchRecord(t, s, "job", Record{StartedAt: now.Add(time.Duration(i) * time.Minute), Success: true})
	}

	results, err := Search(s, SearchQuery{Limit: 3})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results with limit, got %d", len(results))
	}
}

func TestSearch_TextSearch(t *testing.T) {
	s := setupSearchStore(t)
	now := time.Now()
	tags, _ := ParseTags([]string{"env:prod"})
	addSearchRecord(t, s, "deploy", Record{StartedAt: now, Success: true, Tags: tags})
	addSearchRecord(t, s, "backup", Record{StartedAt: now, Success: true})

	results, err := Search(s, SearchQuery{TextSearch: "prod"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 || results[0].JobName != "deploy" {
		t.Errorf("expected deploy job via tag text search, got %+v", results)
	}
}

func TestSearch_Empty(t *testing.T) {
	s := setupSearchStore(t)
	results, err := Search(s, SearchQuery{})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results on empty store, got %d", len(results))
	}
}
