package history

import (
	"testing"
	"time"
)

func addSpikeRecord(t *testing.T, s *Store, job, status string, dur time.Duration) {
	t.Helper()
	err := s.Add(Record{
		JobName:   job,
		Status:    status,
		StartedAt: time.Now(),
		Duration:  dur,
	})
	if err != nil {
		t.Fatalf("addSpikeRecord: %v", err)
	}
}

func TestDetectSpikes_FlagsSpikyRun(t *testing.T) {
	s := New(t.TempDir() + "/spike.db")
	for i := 0; i < 10; i++ {
		addSpikeRecord(t, s, "backup", "success", 2*time.Second)
	}
	// One very slow run — 10× the median.
	addSpikeRecord(t, s, "backup", "success", 20*time.Second)

	results, err := DetectSpikes(s, "backup", SpikeOptions{Multiplier: 3.0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 spike, got %d", len(results))
	}
	if results[0].Duration != 20*time.Second {
		t.Errorf("expected spike duration 20s, got %v", results[0].Duration)
	}
	if results[0].Ratio < 3.0 {
		t.Errorf("expected ratio >= 3.0, got %.2f", results[0].Ratio)
	}
}

func TestDetectSpikes_NoSpikesWhenUniform(t *testing.T) {
	s := New(t.TempDir() + "/spike2.db")
	for i := 0; i < 10; i++ {
		addSpikeRecord(t, s, "sync", "success", 5*time.Second)
	}

	results, err := DetectSpikes(s, "sync", SpikeOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no spikes, got %d", len(results))
	}
}

func TestDetectSpikes_InsufficientSamples(t *testing.T) {
	s := New(t.TempDir() + "/spike3.db")
	addSpikeRecord(t, s, "job", "success", 1*time.Second)

	results, err := DetectSpikes(s, "job", SpikeOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results for insufficient samples")
	}
}

func TestDetectSpikes_MissingJob(t *testing.T) {
	s := New(t.TempDir() + "/spike4.db")

	_, err := DetectSpikes(s, "ghost", SpikeOptions{})
	if err == nil {
		t.Fatal("expected error for missing job")
	}
}

func TestDetectSpikes_IgnoresFailedRuns(t *testing.T) {
	s := New(t.TempDir() + "/spike5.db")
	for i := 0; i < 8; i++ {
		addSpikeRecord(t, s, "etl", "success", 3*time.Second)
	}
	// Failed runs with extreme durations should not influence detection.
	for i := 0; i < 5; i++ {
		addSpikeRecord(t, s, "etl", "failure", 60*time.Second)
	}

	results, err := DetectSpikes(s, "etl", SpikeOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected no spikes (failures ignored), got %d", len(results))
	}
}
