// Package history provides Correlate to measure how closely two cron jobs'
// execution durations move together over time.
//
// # Overview
//
// Correlate computes the Pearson correlation coefficient between the durations
// of two named jobs using their most recent recorded runs. The result ranges
// from -1.0 (perfectly inversely correlated) to +1.0 (perfectly correlated),
// with 0.0 indicating no linear relationship.
//
// # Usage
//
//	res, err := history.Correlate(store, "backup", "cleanup", 30)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Correlation between %s and %s: %.2f (%d samples)\n",
//		res.JobA, res.JobB, res.Coefficient, res.SampleCount)
//
// # Notes
//
// At least two records per job are required. When the two jobs have different
// numbers of records, only the most recent N records (where N is the minimum
// of the two counts, capped at maxSamples) are compared positionally.
package history
