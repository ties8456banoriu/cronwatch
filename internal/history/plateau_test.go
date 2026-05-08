package history

import (
	"testing"
	"time"
)

func addPlateauRecord(t *testing.T, s *Store, job string, durationMs float64, at time.Time) {
	t.Helper()
	err := s.Add(Record{
		JobName:    job,
		Status:     "success",
		DurationMs: durationMs,
		StartedAt:  at,
	})
	if err != nil {
		t.Fatalf("addPlateauRecord: %v", err)
	}
}

func TestDetectPlateau_FindsStableRun(t *testing.T) {
	s := New(t.TempDir() + "/plateau.db")
	now := time.Now()
	for i := 0; i < 5; i++ {
		addPlateauRecord(t, s, "backup", 100.0+float64(i), now.Add(time.Duration(i)*time.Minute))
	}
	res, err := DetectPlateau(s, "backup", PlateauOptions{MinRun: 3, TolerancePct: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Fatal("expected a plateau result, got nil")
	}
	if res.SampleCount < 3 {
		t.Errorf("expected SampleCount >= 3, got %d", res.SampleCount)
	}
	if res.JobName != "backup" {
		t.Errorf("expected job name 'backup', got %q", res.JobName)
	}
}

func TestDetectPlateau_VolatileData_ReturnsNil(t *testing.T) {
	s := New(t.TempDir() + "/plateau2.db")
	now := time.Now()
	durations := []float64{100, 500, 101, 499, 102}
	for i, d := range durations {
		addPlateauRecord(t, s, "volatile", d, now.Add(time.Duration(i)*time.Minute))
	}
	res, err := DetectPlateau(s, "volatile", PlateauOptions{MinRun: 3, TolerancePct: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != nil {
		t.Errorf("expected nil plateau for volatile data, got %+v", res)
	}
}

func TestDetectPlateau_InsufficientSamples(t *testing.T) {
	s := New(t.TempDir() + "/plateau3.db")
	now := time.Now()
	addPlateauRecord(t, s, "rare", 200, now)
	addPlateauRecord(t, s, "rare", 201, now.Add(time.Minute))
	res, err := DetectPlateau(s, "rare", PlateauOptions{MinRun: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != nil {
		t.Errorf("expected nil when fewer than MinRun samples, got %+v", res)
	}
}

func TestDetectPlateau_IgnoresFailedRuns(t *testing.T) {
	s := New(t.TempDir() + "/plateau4.db")
	now := time.Now()
	for i := 0; i < 4; i++ {
		addPlateauRecord(t, s, "mixed", 100, now.Add(time.Duration(i)*time.Minute))
	}
	// inject a failure that should not disrupt plateau detection
	_ = s.Add(Record{JobName: "mixed", Status: "failure", DurationMs: 9999, StartedAt: now.Add(5 * time.Minute)})
	res, err := DetectPlateau(s, "mixed", PlateauOptions{MinRun: 3, TolerancePct: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Fatal("expected plateau result ignoring failed runs")
	}
}

func TestDetectPlateau_NilStore(t *testing.T) {
	_, err := DetectPlateau(nil, "job", PlateauOptions{})
	if err == nil {
		t.Error("expected error for nil store")
	}
}
