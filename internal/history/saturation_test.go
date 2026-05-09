package history

import (
	"testing"
	"time"
)

func addSaturationRecord(t *testing.T, st Store, job, status string, dur time.Duration) {
	t.Helper()
	err := st.Add(Record{
		JobName:   job,
		Status:    status,
		StartedAt: time.Now(),
		Duration:  dur,
	})
	if err != nil {
		t.Fatalf("addSaturationRecord: %v", err)
	}
}

func TestComputeSaturation_UnderBudget(t *testing.T) {
	st := New(t.TempDir() + "/sat.db")
	for i := 0; i < 5; i++ {
		addSaturationRecord(t, st, "job1", "success", 200*time.Millisecond)
	}
	res, err := ComputeSaturation(st, "job1", SaturationOptions{BudgetMs: 1000})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.OverBudget {
		t.Errorf("expected under budget, got over")
	}
	if res.Saturation >= 1.0 {
		t.Errorf("saturation should be < 1.0, got %.3f", res.Saturation)
	}
}

func TestComputeSaturation_OverBudget(t *testing.T) {
	st := New(t.TempDir() + "/sat.db")
	for i := 0; i < 5; i++ {
		addSaturationRecord(t, st, "job2", "success", 1500*time.Millisecond)
	}
	res, err := ComputeSaturation(st, "job2", SaturationOptions{BudgetMs: 1000})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.OverBudget {
		t.Errorf("expected over budget")
	}
	if res.Saturation < 1.0 {
		t.Errorf("saturation should be >= 1.0, got %.3f", res.Saturation)
	}
}

func TestComputeSaturation_IgnoresFailures(t *testing.T) {
	st := New(t.TempDir() + "/sat.db")
	for i := 0; i < 4; i++ {
		addSaturationRecord(t, st, "job3", "failure", 5000*time.Millisecond)
	}
	addSaturationRecord(t, st, "job3", "success", 100*time.Millisecond)
	res, err := ComputeSaturation(st, "job3", SaturationOptions{BudgetMs: 1000})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.OverBudget {
		t.Errorf("failures should be ignored; job should be under budget")
	}
}

func TestComputeSaturation_MissingJob(t *testing.T) {
	st := New(t.TempDir() + "/sat.db")
	_, err := ComputeSaturation(st, "ghost", SaturationOptions{BudgetMs: 500})
	if err == nil {
		t.Fatal("expected error for missing job")
	}
}

func TestComputeSaturation_InvalidBudget(t *testing.T) {
	st := New(t.TempDir() + "/sat.db")
	addSaturationRecord(t, st, "job4", "success", 100*time.Millisecond)
	_, err := ComputeSaturation(st, "job4", SaturationOptions{BudgetMs: 0})
	if err == nil {
		t.Fatal("expected error for zero budget")
	}
}
