package history

import (
	"testing"
	"time"
)

func addBurndownRecord(t *testing.T, st Store, job string, success bool, ago time.Duration) {
	t.Helper()
	r := Record{
		JobName:   job,
		StartedAt: time.Now().Add(-ago),
		Duration:  time.Second,
		Success:   success,
	}
	if err := st.Add(r); err != nil {
		t.Fatalf("addBurndownRecord: %v", err)
	}
}

func TestComputeBurndown_AllSuccesses(t *testing.T) {
	st := New(tempPath(t))
	for i := 0; i < 5; i++ {
		addBurndownRecord(t, st, "job", true, time.Duration(i)*time.Hour)
	}
	res, err := ComputeBurndown(st, "job", BurndownOptions{Window: 7 * 24 * time.Hour, BudgetPercent: 1.0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Exhausted {
		t.Error("expected budget not exhausted")
	}
	for _, pt := range res.Points {
		if pt.Consumed != 0 {
			t.Errorf("expected 0 consumed, got %f", pt.Consumed)
		}
	}
}

func TestComputeBurndown_ExhaustedBudget(t *testing.T) {
	st := New(tempPath(t))
	// 3 failures out of 3 = 100% failure rate, budget 1%
	for i := 0; i < 3; i++ {
		addBurndownRecord(t, st, "job", false, time.Duration(i+1)*time.Hour)
	}
	res, err := ComputeBurndown(st, "job", BurndownOptions{Window: 7 * 24 * time.Hour, BudgetPercent: 1.0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Exhausted {
		t.Error("expected budget to be exhausted")
	}
	if res.ExhaustedAt == nil {
		t.Error("expected ExhaustedAt to be set")
	}
}

func TestComputeBurndown_PointCount(t *testing.T) {
	st := New(tempPath(t))
	for i := 0; i < 10; i++ {
		addBurndownRecord(t, st, "job", true, time.Duration(i)*time.Hour)
	}
	res, err := ComputeBurndown(st, "job", BurndownOptions{Window: 30 * 24 * time.Hour, BudgetPercent: 5.0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Points) != 10 {
		t.Errorf("expected 10 points, got %d", len(res.Points))
	}
}

func TestComputeBurndown_MissingJob(t *testing.T) {
	st := New(tempPath(t))
	_, err := ComputeBurndown(st, "ghost", BurndownOptions{})
	if err == nil {
		t.Error("expected error for missing job")
	}
}

func TestComputeBurndown_NilStore(t *testing.T) {
	_, err := ComputeBurndown(nil, "job", BurndownOptions{})
	if err == nil {
		t.Error("expected error for nil store")
	}
}
