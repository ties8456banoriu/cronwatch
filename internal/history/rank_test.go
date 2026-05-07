package history

import (
	"testing"
	"time"
)

func addRankRecord(t *testing.T, s *Store, job string, dur time.Duration, success bool) {
	t.Helper()
	status := "success"
	if !success {
		status = "failure"
	}
	err := s.Add(Record{
		JobName:   job,
		StartedAt: time.Now(),
		Duration:  dur,
		Status:    status,
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
}

func TestRankJobs_OrderedBestFirst(t *testing.T) {
	s := New(t.TempDir() + "/rank.json")

	// "fast" job: all successes, short duration → high score
	for i := 0; i < 10; i++ {
		addRankRecord(t, s, "fast", 100*time.Millisecond, true)
	}
	// "slow" job: all successes but very slow → lower score
	for i := 0; i < 10; i++ {
		addRankRecord(t, s, "slow", 10*time.Second, true)
	}
	// "flaky" job: many failures → lowest score
	for i := 0; i < 5; i++ {
		addRankRecord(t, s, "flaky", 200*time.Millisecond, false)
	}
	addRankRecord(t, s, "flaky", 200*time.Millisecond, true)

	ranks, err := RankJobs(s, RankOptions{})
	if err != nil {
		t.Fatalf("RankJobs: %v", err)
	}
	if len(ranks) != 3 {
		t.Fatalf("expected 3 ranked jobs, got %d", len(ranks))
	}
	if ranks[0].JobName != "fast" {
		t.Errorf("expected 'fast' to rank first, got %q", ranks[0].JobName)
	}
	if ranks[len(ranks)-1].JobName != "flaky" {
		t.Errorf("expected 'flaky' to rank last, got %q", ranks[len(ranks)-1].JobName)
	}
	for i, r := range ranks {
		if r.Rank != i+1 {
			t.Errorf("rank[%d].Rank = %d, want %d", i, r.Rank, i+1)
		}
	}
}

func TestRankJobs_Ascending(t *testing.T) {
	s := New(t.TempDir() + "/rank_asc.json")

	for i := 0; i < 8; i++ {
		addRankRecord(t, s, "good", 50*time.Millisecond, true)
	}
	for i := 0; i < 4; i++ {
		addRankRecord(t, s, "bad", 100*time.Millisecond, false)
	}
	addRankRecord(t, s, "bad", 100*time.Millisecond, true)

	ranks, err := RankJobs(s, RankOptions{Ascending: true})
	if err != nil {
		t.Fatalf("RankJobs: %v", err)
	}
	if ranks[0].JobName != "bad" {
		t.Errorf("ascending: expected 'bad' first, got %q", ranks[0].JobName)
	}
}

func TestRankJobs_EmptyStore(t *testing.T) {
	s := New(t.TempDir() + "/rank_empty.json")
	ranks, err := RankJobs(s, RankOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ranks) != 0 {
		t.Errorf("expected empty result, got %d entries", len(ranks))
	}
}

func TestRankJobs_NilStore(t *testing.T) {
	_, err := RankJobs(nil, RankOptions{})
	if err == nil {
		t.Error("expected error for nil store, got nil")
	}
}
