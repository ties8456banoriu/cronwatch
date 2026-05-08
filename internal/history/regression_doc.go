// Package history provides regression analysis for cron job run durations.
//
// # Linear Regression
//
// ComputeRegression fits a simple ordinary-least-squares line to the
// successful run durations of a named job and returns:
//
//   - Slope     — how many milliseconds the duration grows per run
//   - Intercept — estimated duration at run index 0
//   - R²        — goodness-of-fit (0 = no fit, 1 = perfect fit)
//   - Warning   — non-empty string when slope exceeds the configured threshold
//
// # Usage
//
//	res, err := history.ComputeRegression(store, "nightly-backup", history.RegressionOptions{
//		MaxSamples:         50,
//		SlopeWarnThreshold: 10, // alert if growing >10 ms per run
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	if res.Warning != "" {
//		notifier.Send(context.Background(), alert.Alert{Message: res.Warning})
//	}
//
// A minimum of 3 successful samples is required; fewer returns an error.
package history
