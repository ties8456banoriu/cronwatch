// Package history provides plateau detection via DetectPlateau.
//
// # Plateau Detection
//
// A "plateau" is a consecutive sequence of successful job runs whose execution
// durations remain within a narrow band. Detecting a plateau indicates that a
// job has reached a stable performance regime — useful for setting dynamic
// baselines or suppressing noise-driven alerts.
//
// # Usage
//
//	res, err := history.DetectPlateau(store, "nightly-backup", history.PlateauOptions{
//		MinRun:       5,
//		TolerancePct: 8.0,
//		MaxSamples:   50,
//	})
//	if res != nil {
//		fmt.Printf("Stable for %d runs, avg %.1f ms\n", res.SampleCount, res.AvgDurationMs)
//	}
//
// # Options
//
//   - MinRun: minimum consecutive samples to declare a plateau (default 3).
//   - TolerancePct: maximum (max-min)/avg spread as a percentage (default 10).
//   - MaxSamples: cap on how many recent records are examined (0 = unlimited).
package history
