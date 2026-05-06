package monitor

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/cronwatch/cronwatch/internal/alert"
	"github.com/cronwatch/cronwatch/internal/history"
)

// BaselineChecker compares incoming records against stored baselines and
// sends an alert when a job's duration drifts outside the allowed tolerance.
type BaselineChecker struct {
	baselineDir string
	notifier    alert.Notifier
	logger      *log.Logger
}

// NewBaselineChecker creates a BaselineChecker that loads baselines from dir.
func NewBaselineChecker(baselineDir string, notifier alert.Notifier) *BaselineChecker {
	return &BaselineChecker{
		baselineDir: baselineDir,
		notifier:    notifier,
		logger:      log.New(os.Stderr, "[baseline] ", log.LstdFlags),
	}
}

// Check loads the baseline for the record's job and compares durations.
// If no baseline file exists the check is silently skipped.
func (bc *BaselineChecker) Check(r history.Record) error {
	b, err := history.LoadBaseline(bc.baselineDir, r.JobName)
	if err != nil {
		if isNotFound(bc.baselineDir, r.JobName) {
			return nil // no baseline established yet
		}
		return fmt.Errorf("baseline check: load: %w", err)
	}

	res := history.CompareToBaseline(r, b)
	if res.Within {
		return nil
	}

	direction := "slower"
	if res.Delta < 0 {
		direction = "faster"
	}
	msg := fmt.Sprintf(
		"[cronwatch] job %q ran %.0f%% %s than baseline (expected %v, got %v)",
		r.JobName,
		res.DeltaPct*100,
		direction,
		b.ExpectedDuration,
		r.Duration,
	)
	bc.logger.Println(msg)
	return bc.notifier.Send(msg)
}

// isNotFound returns true when the expected baseline file does not exist.
func isNotFound(dir, jobName string) bool {
	path := filepath.Join(dir, jobName+"_baseline.json")
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}
