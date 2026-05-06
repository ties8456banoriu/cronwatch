package monitor_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/history"
	"github.com/cronwatch/cronwatch/internal/monitor"
)

type captureNotifier struct {
	msgs []string
}

func (c *captureNotifier) Send(msg string) error {
	c.msgs = append(c.msgs, msg)
	return nil
}

type failNotifier struct{}

func (f *failNotifier) Send(_ string) error {
	return fmt.Errorf("notifier failure")
}

func TestBaselineChecker_WithinTolerance_NoAlert(t *testing.T) {
	dir := t.TempDir()
	b := history.Baseline{
		JobName:         "job",
		ExpectedDuration: 10 * time.Second,
		TolerancePct:    0.20,
	}
	if err := history.SaveBaseline(dir, b); err != nil {
		t.Fatalf("save baseline: %v", err)
	}

	n := &captureNotifier{}
	bc := monitor.NewBaselineChecker(dir, n)

	r := history.Record{JobName: "job", Duration: 11 * time.Second, Success: true}
	if err := bc.Check(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(n.msgs) != 0 {
		t.Errorf("expected no alert, got: %v", n.msgs)
	}
}

func TestBaselineChecker_OutsideTolerance_SendsAlert(t *testing.T) {
	dir := t.TempDir()
	b := history.Baseline{
		JobName:         "slowjob",
		ExpectedDuration: 10 * time.Second,
		TolerancePct:    0.10,
	}
	if err := history.SaveBaseline(dir, b); err != nil {
		t.Fatalf("save baseline: %v", err)
	}

	n := &captureNotifier{}
	bc := monitor.NewBaselineChecker(dir, n)

	r := history.Record{JobName: "slowjob", Duration: 20 * time.Second, Success: true}
	if err := bc.Check(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(n.msgs) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(n.msgs))
	}
	if n.msgs[0] == "" {
		t.Error("alert message should not be empty")
	}
}

func TestBaselineChecker_NoBaseline_SkipsSilently(t *testing.T) {
	dir := t.TempDir()
	n := &captureNotifier{}
	bc := monitor.NewBaselineChecker(dir, n)

	r := history.Record{JobName: "unknown", Duration: 5 * time.Second, Success: true}
	if err := bc.Check(r); err != nil {
		t.Fatalf("expected no error for missing baseline, got: %v", err)
	}
	if len(n.msgs) != 0 {
		t.Errorf("expected no alert, got: %v", n.msgs)
	}
}

func TestBaselineChecker_NotifierError_Propagates(t *testing.T) {
	dir := t.TempDir()
	b := history.Baseline{
		JobName:         "eryjob",
		ExpectedDuration: 5 * time.Second,
		TolerancePct:    0.05,
	}
	if err := history.SaveBaseline(dir, b); err != nil {
		t.Fatalf("save baseline: %v", err)
	}

	bc := monitor.NewBaselineChecker(dir, &failNotifier{})
	r := history.Record{JobName: "eryjob", Duration: 30 * time.Second, Success: true}
	if err := bc.Check(r); err == nil {
		t.Fatal("expected notifier error to propagate")
	}
}
