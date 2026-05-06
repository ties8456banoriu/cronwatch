// Package history provides run-history storage and analysis for cronwatch.
//
// # Cluster
//
// Cluster groups a job's historical run durations into K clusters using the
// k-means algorithm (Lloyd's method). This is useful for detecting multimodal
// duration distributions — for example, a backup job that sometimes runs fast
// (incremental) and sometimes runs slow (full).
//
// Only successful records are considered. The centroids are seeded evenly
// across the sorted duration range, which avoids the pathological empty-cluster
// problem for well-separated distributions.
//
// Example:
//
//	res, err := history.Cluster(store, "backup", history.ClusterOptions{
//		K:          2,
//		MaxIter:    30,
//		MaxSamples: 200,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("centroids: %v\n", res.Centroids)
//	fmt.Printf("inertia:   %.2f\n", res.Inertia)
package history
