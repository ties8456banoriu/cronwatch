package history

import (
	"errors"
	"testing"
	"time"
)

func addReplayRecord(t *testing.T, s *Store, job string, start time.Time, success bool) {
	t.Helper()
	r := Record{
		JobName:   job,
		StartedAt: start,
		Duration:  time.Second,
		Success:   success,
	}
	if err := s.Add(r); err != nil {
		t.Fatalf("add record: %v", err)
	}
}

func TestReplay_DryRun_NoSideEffects(t *testing.T) {
	s := New(t.TempDir() + "/replay.db")
	now := time.Now()
	addReplayRecord(t, s, "job1", now.Add(-2*time.Hour), true)
	addReplayRecord(t, s, "job1", now.Add(-1*time.Hour), true)

	called := 0
	results, err := Replay(s, ReplayOptions{JobName: "job1", DryRun: true}, func(r Record) error {
		called++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called != 0 {
		t.Errorf("fn should not be called in dry-run, got %d calls", called)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestReplay_CallsFnForEachRecord(t *testing.T) {
	s := New(t.TempDir() + "/replay2.db")
	now := time.Now()
	addReplayRecord(t, s, "job2", now.Add(-3*time.Hour), true)
	addReplayRecord(t, s, "job2", now.Add(-2*time.Hour), false)

	var seen []Record
	_, err := Replay(s, ReplayOptions{JobName: "job2"}, func(r Record) error {
		seen = append(seen, r)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(seen) != 2 {
		t.Errorf("expected 2 records replayed, got %d", len(seen))
	}
}

func TestReplay_SinceFilter(t *testing.T) {
	s := New(t.TempDir() + "/replay3.db")
	now := time.Now()
	addReplayRecord(t, s, "job3", now.Add(-5*time.Hour), true)
	addReplayRecord(t, s, "job3", now.Add(-1*time.Hour), true)

	var seen []Record
	_, err := Replay(s, ReplayOptions{JobName: "job3", Since: now.Add(-2 * time.Hour)}, func(r Record) error {
		seen = append(seen, r)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(seen) != 1 {
		t.Errorf("expected 1 record after Since filter, got %d", len(seen))
	}
}

func TestReplay_FnError_Propagates(t *testing.T) {
	s := New(t.TempDir() + "/replay4.db")
	addReplayRecord(t, s, "job4", time.Now().Add(-time.Hour), true)

	sentinel := errors.New("fn failed")
	_, err := Replay(s, ReplayOptions{JobName: "job4"}, func(r Record) error {
		return sentinel
	})
	if err == nil {
		t.Fatal("expected error from fn, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestReplay_EmptyJobName_ReturnsError(t *testing.T) {
	s := New(t.TempDir() + "/replay5.db")
	_, err := Replay(s, ReplayOptions{}, func(r Record) error { return nil })
	if err == nil {
		t.Fatal("expected error for empty JobName")
	}
}
