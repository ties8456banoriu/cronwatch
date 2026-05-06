package history

import (
	"errors"
	"math"
	"sort"
	"time"
)

// ClusterResult holds the output of a duration clustering operation for a job.
type ClusterResult struct {
	JobName  string
	Centroids []time.Duration // one centroid per cluster
	Labels    []int           // cluster assignment per record (same order as records)
	Inertia   float64         // sum of squared distances to assigned centroid
}

// ClusterOptions controls k-means clustering behaviour.
type ClusterOptions struct {
	// K is the number of clusters. Must be >= 2.
	K int
	// MaxIter caps the number of Lloyd's iterations.
	MaxIter int
	// MaxSamples limits how many recent successful records are used.
	MaxSamples int
}

// Cluster groups a job's run durations into K clusters using k-means.
// Only successful records are considered. Returns ErrNotFound if the job has
// no history, and ErrInsufficientData if there are fewer records than K.
func Cluster(s *Store, jobName string, opts ClusterOptions) (*ClusterResult, error) {
	if opts.K < 2 {
		return nil, errors.New("cluster: K must be >= 2")
	}
	if opts.MaxIter <= 0 {
		opts.MaxIter = 50
	}

	recs, err := s.All(jobName)
	if err != nil {
		return nil, err
	}
	if len(recs) == 0 {
		return nil, ErrNotFound
	}

	// Filter to successful runs only.
	var durations []float64
	for _, r := range recs {
		if r.Success {
			durations = append(durations, float64(r.Duration))
		}
	}
	if opts.MaxSamples > 0 && len(durations) > opts.MaxSamples {
		durations = durations[len(durations)-opts.MaxSamples:]
	}
	if len(durations) < opts.K {
		return nil, ErrInsufficientData
	}

	// Seed centroids evenly across the sorted range.
	sorted := make([]float64, len(durations))
	copy(sorted, durations)
	sort.Float64s(sorted)

	centroids := make([]float64, opts.K)
	step := (len(sorted) - 1) / (opts.K - 1)
	for i := 0; i < opts.K; i++ {
		centroids[i] = sorted[i*step]
	}

	labels := make([]int, len(durations))
	for iter := 0; iter < opts.MaxIter; iter++ {
		// Assignment step.
		changed := false
		for i, d := range durations {
			best, bestDist := 0, math.MaxFloat64
			for k, c := range centroids {
				if dist := math.Abs(d - c); dist < bestDist {
					best, bestDist = k, dist
				}
			}
			if labels[i] != best {
				labels[i] = best
				changed = true
			}
		}
		if !changed {
			break
		}
		// Update step.
		sums := make([]float64, opts.K)
		counts := make([]int, opts.K)
		for i, d := range durations {
			sums[labels[i]] += d
			counts[labels[i]]++
		}
		for k := range centroids {
			if counts[k] > 0 {
				centroids[k] = sums[k] / float64(counts[k])
			}
		}
	}

	// Compute inertia.
	var inertia float64
	for i, d := range durations {
		diff := d - centroids[labels[i]]
		inertia += diff * diff
	}

	resultCentroids := make([]time.Duration, opts.K)
	for k, c := range centroids {
		resultCentroids[k] = time.Duration(c)
	}

	return &ClusterResult{
		JobName:   jobName,
		Centroids: resultCentroids,
		Labels:    labels,
		Inertia:   inertia,
	}, nil
}
