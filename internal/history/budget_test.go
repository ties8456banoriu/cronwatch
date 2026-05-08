package history

import (
	"testing"
	"time"
)

func addBudgetRecord(t *testing.T, s *Store, job string, success bool, ago time.Duration) {
	t.Helper()
	err := s.Add(Record{
		JobName:   job,
		StartedAt: time.Now().Add(-ago),
		Duration:  100 * time.Millisecond,
		Success:   success,
	})
	if err != nil {
		t.Fatalf("addBudgetRecord: %v", err)
	}
}

func TestComputeBudget_PerfectRecord(t *testing.T) {
	s := New(t.TempDir() + "/budget.db")
	for i := 0; i < 10; i++ {
		addBudgetRecord(t, s, "job", true, time.Duration(i)*time.Minute)
	}
	res, err := ComputeBudget(s, "job", 0.99, BudgetOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.FailedRuns != 0 {
		t.Errorf("expected 0 failures, got %d", res.FailedRuns)
	}
	if res.BudgetConsumed != 0 {
		t.Errorf("expected 0 consumed, got %f", res.BudgetConsumed)
	}
	if res.Exhausted {
		t.Error("expected budget not exhausted")
	}
}

func TestComputeBudget_ExhaustedBudget(t *testing.T) {
	s := New(t.TempDir() + "/budget.db")
	// 5 failures out of 10 runs, SLO 0.99 allows 0.1 failures → exhausted
	for i := 0; i < 5; i++ {
		addBudgetRecord(t, s, "job", false, time.Duration(i)*time.Minute)
	}
	for i := 5; i < 10; i++ {
		addBudgetRecord(t, s, "job", true, time.Duration(i)*time.Minute)
	}
	res, err := ComputeBudget(s, "job", 0.99, BudgetOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Exhausted {
		t.Error("expected budget to be exhausted")
	}
	if res.BudgetConsumed != 1.0 {
		t.Errorf("expected consumed=1.0, got %f", res.BudgetConsumed)
	}
	if res.BudgetRemaining != 0 {
		t.Errorf("expected remaining=0, got %f", res.BudgetRemaining)
	}
}

func TestComputeBudget_WindowFilters(t *testing.T) {
	s := New(t.TempDir() + "/budget.db")
	// old failure — outside window
	addBudgetRecord(t, s, "job", false, 3*24*time.Hour)
	// recent successes
	for i := 0; i < 5; i++ {
		addBudgetRecord(t, s, "job", true, time.Duration(i)*time.Minute)
	}
	res, err := ComputeBudget(s, "job", 0.95, BudgetOptions{Window: 24 * time.Hour})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.FailedRuns != 0 {
		t.Errorf("expected 0 failures in window, got %d", res.FailedRuns)
	}
}

func TestComputeBudget_MissingJob(t *testing.T) {
	s := New(t.TempDir() + "/budget.db")
	_, err := ComputeBudget(s, "ghost", 0.99, BudgetOptions{})
	if err == nil {
		t.Error("expected error for missing job")
	}
}

func TestComputeBudget_InvalidSLO(t *testing.T) {
	s := New(t.TempDir() + "/budget.db")
	addBudgetRecord(t, s, "job", true, time.Minute)
	_, err := ComputeBudget(s, "job", 1.5, BudgetOptions{})
	if err == nil {
		t.Error("expected error for SLO > 1")
	}
	_, err = ComputeBudget(s, "job", 0, BudgetOptions{})
	if err == nil {
		t.Error("expected error for SLO == 0")
	}
}

func TestComputeBudget_NilStore(t *testing.T) {
	_, err := ComputeBudget(nil, "job", 0.99, BudgetOptions{})
	if err == nil {
		t.Error("expected error for nil store")
	}
}
