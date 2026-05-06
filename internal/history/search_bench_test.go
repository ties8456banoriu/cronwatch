package history

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkSearch_LargeStore(b *testing.B) {
	s, err := New(b.TempDir() + "/bench.db")
	if err != nil {
		b.Fatalf("New: %v", err)
	}
	defer s.Close()

	now := time.Now()
	for j := 0; j < 10; j++ {
		name := fmt.Sprintf("job-%d", j)
		for i := 0; i < 50; i++ {
			_ = s.Add(name, Record{
				StartedAt: now.Add(-time.Duration(i) * time.Minute),
				Duration:  time.Second * time.Duration(i%10+1),
				Success:   i%3 != 0,
			})
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Search(s, SearchQuery{Status: "failure", Limit: 20})
		if err != nil {
			b.Fatalf("Search: %v", err)
		}
	}
}

func BenchmarkSearch_ByJobName(b *testing.B) {
	s, err := New(b.TempDir() + "/bench2.db")
	if err != nil {
		b.Fatalf("New: %v", err)
	}
	defer s.Close()

	now := time.Now()
	for i := 0; i < 100; i++ {
		_ = s.Add("target-job", Record{
			StartedAt: now.Add(-time.Duration(i) * time.Minute),
			Duration:  time.Second,
			Success:   true,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Search(s, SearchQuery{JobName: "target-job"})
		if err != nil {
			b.Fatalf("Search: %v", err)
		}
	}
}
