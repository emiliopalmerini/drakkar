package search

import "context"

// FindRequest parameters for a memory find operation.
type FindRequest struct {
	Query          string
	Limit          int
	ScoreThreshold float64
	TargetURI      string
}

// SearchRequest parameters for a context search operation.
type SearchRequest struct {
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
	Find(ctx context.Context, req FindRequest) (*FindResult, error)
	Search(ctx context.Context, req SearchRequest) (*FindResult, error)
}
