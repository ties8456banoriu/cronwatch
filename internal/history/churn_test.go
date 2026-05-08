package history

import (
	"testing"
	"time"
)

func addChurnRecord(t *testing.T, st *Store, job, status string, at time.Time) {
	t.Helper()
	err := st.Add(Record{
		JobName:   job,
		Status:    status,
		StartedAt: at,
		Duration:  100 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("addChurnRecord: %v", err)
	}
}

func TestComputeChurn_FrequentFlips(t *testing.T) {
	st := New(t.TempDir() + "/churn.json")
	now := time.Now()
	statuses := []string{"success", "failure", "success", "failure", "success", "failure"}
	for i, s := range statuses {
		addChurnRecord(t, st, "flipper", s, now.Add(-time.Duration(len(statuses)-i)*time.Minute))
	}

	res, err := ComputeChurn(st, "flipper", ChurnOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Transitions != 5 {
		t.Errorf("expected 5 transitions, got %d", res.Transitions)
	}
	if !res.IsChurning {
		t.Error("expected IsChurning=true")
	}
}

func TestComputeChurn_StableJob_NoChurn(t *testing.T) {
	st := New(t.TempDir() + "/churn.json")
	now := time.Now()
	for i := 0; i < 6; i++ {
		addChurnRecord(t, st, "stable", "success", now.Add(-time.Duration(6-i)*time.Minute))
	}

	res, err := ComputeChurn(st, "stable", ChurnOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Transitions != 0 {
		t.Errorf("expected 0 transitions, got %d", res.Transitions)
	}
	if res.IsChurning {
		t.Error("expected IsChurning=false")
	}
}

func TestComputeChurn_WindowFilters(t *testing.T) {
	st := New(t.TempDir() + "/churn.json")
	now := time.Now()
	// Old alternating records — outside window.
	for i := 0; i < 4; i++ {
		status := "success"
		if i%2 == 1 {
			status = "failure"
		}
		addChurnRecord(t, st, "job", status, now.Add(-48*time.Hour-time.Duration(i)*time.Minute))
	}
	// Recent stable records — inside window.
	for i := 0; i < 4; i++ {
		addChurnRecord(t, st, "job", "success", now.Add(-time.Duration(i)*time.Minute))
	}

	res, err := ComputeChurn(st, "job", ChurnOptions{Window: 24 * time.Hour})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Transitions != 0 {
		t.Errorf("expected 0 transitions within window, got %d", res.Transitions)
	}
}

func TestComputeChurn_MissingJob(t *testing.T) {
	st := New(t.TempDir() + "/churn.json")
	_, err := ComputeChurn(st, "ghost", ChurnOptions{})
	if err == nil {
		t.Error("expected error for missing job")
	}
}

func TestComputeChurn_NilStore(t *testing.T) {
	_, err := ComputeChurn(nil, "job", ChurnOptions{})
	if err == nil {
		t.Error("expected error for nil store")
	}
}
