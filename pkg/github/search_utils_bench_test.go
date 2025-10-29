package github

import (
	"testing"
)

// BenchmarkHasFilter benchmarks the hasFilter function with different query types
func BenchmarkHasFilter(b *testing.B) {
	benchmarks := []struct {
		name       string
		query      string
		filterType string
	}{
		{
			name:       "simple query with filter",
			query:      "is:issue repo:github/github",
			filterType: "repo",
		},
		{
			name:       "simple query without filter",
			query:      "some search query without the filter",
			filterType: "repo",
		},
		{
			name:       "complex query with multiple filters",
			query:      "is:issue is:open author:user repo:org/repo label:bug",
			filterType: "label",
		},
		{
			name:       "long query with filter at end",
			query:      "this is a very long search query with many words and finally repo:owner/name",
			filterType: "repo",
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				hasFilter(bm.query, bm.filterType)
			}
		})
	}
}

// BenchmarkHasSpecificFilter benchmarks the hasSpecificFilter function
func BenchmarkHasSpecificFilter(b *testing.B) {
	benchmarks := []struct {
		name        string
		query       string
		filterType  string
		filterValue string
	}{
		{
			name:        "simple query with specific filter",
			query:       "is:issue repo:github/github",
			filterType:  "is",
			filterValue: "issue",
		},
		{
			name:        "simple query without specific filter",
			query:       "is:pr repo:github/github",
			filterType:  "is",
			filterValue: "issue",
		},
		{
			name:        "complex query with specific filter",
			query:       "is:issue is:open author:user repo:org/repo label:bug",
			filterType:  "is",
			filterValue: "open",
		},
		{
			name:        "long query with specific filter at end",
			query:       "this is a very long search query with many words and finally is:closed",
			filterType:  "is",
			filterValue: "closed",
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				hasSpecificFilter(bm.query, bm.filterType, bm.filterValue)
			}
		})
	}
}

// BenchmarkHasRepoFilter benchmarks the hasRepoFilter convenience function
func BenchmarkHasRepoFilter(b *testing.B) {
	queries := []string{
		"is:issue repo:github/github",
		"simple query without repo filter",
		"repo:org/name is:pr author:user",
	}

	for _, query := range queries {
		b.Run(query, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				hasRepoFilter(query)
			}
		})
	}
}

// BenchmarkHasTypeFilter benchmarks the hasTypeFilter convenience function
func BenchmarkHasTypeFilter(b *testing.B) {
	queries := []string{
		"type:user some search",
		"simple query without type filter",
		"type:org location:seattle",
	}

	for _, query := range queries {
		b.Run(query, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				hasTypeFilter(query)
			}
		})
	}
}
