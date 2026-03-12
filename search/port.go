package search

import "context"

// Request parameters for a find or search operation.
type Request struct {
	Query          string
	Limit          int
	ScoreThreshold float64
	TargetURI      string
}

// Result is a single item returned by a search or find operation.
type Result struct {
	URI   string  `json:"uri"`
	Title string  `json:"title"`
	Score float64 `json:"score"`
}

// FindResult is the response from a search or find operation.
type FindResult struct {
	Results []Result `json:"results"`
	Total   int      `json:"total"`
}

// Searcher is the port interface for search and find operations.
// Implementations must be safe for concurrent use.
type Searcher interface {
	Find(ctx context.Context, req Request) (*FindResult, error)
	Search(ctx context.Context, req Request) (*FindResult, error)
}
