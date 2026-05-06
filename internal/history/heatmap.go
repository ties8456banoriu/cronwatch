package history

import (
	"fmt"
	"time"
)

// HeatmapCell represents a single time bucket in the heatmap.
type HeatmapCell struct {
	BucketStart time.Time
	BucketEnd   time.Time
	Count       int
	Failures    int
	AvgDuration time.Duration
}

// Heatmap holds a grid of cells for a single job over a time range.
type Heatmap struct {
	JobName  string
	Granule  time.Duration
	Cells    []HeatmapCell
}

// HeatmapOptions controls how the heatmap is generated.
type HeatmapOptions struct {
	// Granule is the size of each time bucket (e.g. 1*time.Hour).
	Granule time.Duration
	// Since limits records to those after this time. Zero means no lower bound.
	Since time.Time
	// Until limits records to those before this time. Zero means now.
	Until time.Time
}

// BuildHeatmap produces a heatmap of run frequency and health for the given job.
// Returns an error if the job has no records or the options are invalid.
func BuildHeatmap(s *Store, jobName string, opts HeatmapOptions) (*Heatmap, error) {
	if opts.Granule <= 0 {
		return nil, fmt.Errorf("heatmap: granule must be positive")
	}
	until := opts.Until
	if until.IsZero() {
		until = time.Now()
	}

	records, err := s.All(jobName)
	if err != nil {
		return nil, fmt.Errorf("heatmap: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("heatmap: no records found for job %q", jobName)
	}

	// Determine start boundary.
	start := opts.Since
	if start.IsZero() {
		start = records[0].StartedAt.Truncate(opts.Granule)
	} else {
		start = start.Truncate(opts.Granule)
	}

	// Build bucket index.
	buckets := map[time.Time]*HeatmapCell{}
	for t := start; !t.After(until); t = t.Add(opts.Granule) {
		tCopy := t
		buckets[tCopy] = &HeatmapCell{
			BucketStart: tCopy,
			BucketEnd:   tCopy.Add(opts.Granule),
		}
	}

	for _, r := range records {
		if r.StartedAt.Before(start) || r.StartedAt.After(until) {
			continue
		}
		key := r.StartedAt.Truncate(opts.Granule)
		cell, ok := buckets[key]
		if !ok {
			continue
		}
		cell.Count++
		if r.Status == "failed" {
			cell.Failures++
		}
		cell.AvgDuration += r.Duration
	}

	// Finalise averages and collect ordered cells.
	cells := make([]HeatmapCell, 0, len(buckets))
	for t := start; !t.After(until); t = t.Add(opts.Granule) {
		cell := buckets[t]
		if cell.Count > 0 {
			cell.AvgDuration /= time.Duration(cell.Count)
		}
		cells = append(cells, *cell)
	}

	return &Heatmap{
		JobName: jobName,
		Granule: opts.Granule,
		Cells:   cells,
	}, nil
}
