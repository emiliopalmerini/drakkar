package search_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/epalmerini/drakkar/search"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/mcptest"
)

// fakeSearcher is a configurable fake implementation of the Searcher interface.
type fakeSearcher struct {
	findResult   *search.FindResult
	findErr      error
	searchResult *search.FindResult
	searchErr    error

	// Capture last call arguments for assertion.
	lastFindReq   search.FindRequest
	lastSearchReq search.SearchRequest
}

func (f *fakeSearcher) Find(ctx context.Context, req search.FindRequest) (*search.FindResult, error) {
	f.lastFindReq = req
	return f.findResult, f.findErr
}

func (f *fakeSearcher) Search(ctx context.Context, req search.SearchRequest) (*search.FindResult, error) {
	f.lastSearchReq = req
	return f.searchResult, f.searchErr
}

// buildServer registers the search tools against a fake searcher and returns a
// started mcptest server ready to call.
func buildServer(t *testing.T, fake *fakeSearcher) *mcptest.Server {
	t.Helper()
	srv := mcptest.NewUnstartedServer(t)
	search.Register(srv, fake)
	if err := srv.Start(context.Background()); err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	t.Cleanup(srv.Close)
	return srv
}

// callTool invokes a named tool with an arguments map.
func callTool(t *testing.T, s *mcptest.Server, toolName string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Name = toolName
	req.Params.Arguments = args
	result, err := s.Client().CallTool(context.Background(), req)
	if err != nil {
		t.Fatalf("CallTool(%q) error: %v", toolName, err)
	}
	return result
}

// decodeText unmarshals the first text content from a result into dst.
func decodeText(t *testing.T, result *mcp.CallToolResult, dst any) {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("result has no content")
	}
	text, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if err := json.Unmarshal([]byte(text.Text), dst); err != nil {
		t.Fatalf("unmarshal result text: %v", err)
	}
}

// ---- search_memories tests ----

func TestSearchMemories_AllParams(t *testing.T) {
	fake := &fakeSearcher{
		findResult: &search.FindResult{
			Results: []search.Result{
				{URI: "mem://1", Title: "First", Score: 0.95},
				{URI: "mem://2", Title: "Second", Score: 0.87},
			},
			Total: 2,
		},
	}
	s := buildServer(t, fake)

	result := callTool(t, s, "search_memories", map[string]any{
		"query":           "golang context",
		"limit":           5,
		"score_threshold": 0.8,
		"target_uri":      "mem://workspace/go",
	})

	if result.IsError {
		t.Fatalf("expected success, got error result")
	}

	var got search.FindResult
	decodeText(t, result, &got)

	if got.Total != 2 {
		t.Errorf("Total: got %d, want 2", got.Total)
	}
	if len(got.Results) != 2 {
		t.Errorf("Results len: got %d, want 2", len(got.Results))
	}
	if got.Results[0].URI != "mem://1" {
		t.Errorf("Results[0].URI: got %q, want %q", got.Results[0].URI, "mem://1")
	}

	if fake.lastFindReq.Query != "golang context" {
		t.Errorf("Find Query: got %q, want %q", fake.lastFindReq.Query, "golang context")
	}
	if fake.lastFindReq.Limit != 5 {
		t.Errorf("Find Limit: got %d, want 5", fake.lastFindReq.Limit)
	}
	if fake.lastFindReq.ScoreThreshold != 0.8 {
		t.Errorf("Find ScoreThreshold: got %f, want 0.8", fake.lastFindReq.ScoreThreshold)
	}
	if fake.lastFindReq.TargetURI != "mem://workspace/go" {
		t.Errorf("Find TargetURI: got %q, want %q", fake.lastFindReq.TargetURI, "mem://workspace/go")
	}
}

func TestSearchMemories_DefaultParams(t *testing.T) {
	fake := &fakeSearcher{
		findResult: &search.FindResult{Results: []search.Result{}, Total: 0},
	}
	s := buildServer(t, fake)

	result := callTool(t, s, "search_memories", map[string]any{
		"query": "minimal",
	})

	if result.IsError {
		t.Fatalf("expected success, got error result")
	}

	if fake.lastFindReq.Limit != 10 {
		t.Errorf("default Limit: got %d, want 10", fake.lastFindReq.Limit)
	}
	if fake.lastFindReq.ScoreThreshold != 0.0 {
		t.Errorf("default ScoreThreshold: got %f, want 0.0", fake.lastFindReq.ScoreThreshold)
	}
	if fake.lastFindReq.TargetURI != "" {
		t.Errorf("default TargetURI: got %q, want empty", fake.lastFindReq.TargetURI)
	}
}

func TestSearchMemories_EmptyQueryReturnsError(t *testing.T) {
	fake := &fakeSearcher{}
	s := buildServer(t, fake)

	result := callTool(t, s, "search_memories", map[string]any{
		"query": "",
	})

	if !result.IsError {
		t.Fatal("expected error result for empty query")
	}
}

// ---- context_search tests ----

func TestContextSearch_AllParams(t *testing.T) {
	fake := &fakeSearcher{
		searchResult: &search.FindResult{
			Results: []search.Result{
				{URI: "doc://readme", Title: "README", Score: 0.99},
			},
			Total: 1,
		},
	}
	s := buildServer(t, fake)

	result := callTool(t, s, "context_search", map[string]any{
		"query":           "project structure",
		"limit":           3,
		"score_threshold": 0.9,
		"target_uri":      "doc://project",
	})

	if result.IsError {
		t.Fatalf("expected success, got error result")
	}

	var got search.FindResult
	decodeText(t, result, &got)

	if got.Total != 1 {
		t.Errorf("Total: got %d, want 1", got.Total)
	}

	if fake.lastSearchReq.Query != "project structure" {
		t.Errorf("Search Query: got %q, want %q", fake.lastSearchReq.Query, "project structure")
	}
	if fake.lastSearchReq.Limit != 3 {
		t.Errorf("Search Limit: got %d, want 3", fake.lastSearchReq.Limit)
	}
	if fake.lastSearchReq.ScoreThreshold != 0.9 {
		t.Errorf("Search ScoreThreshold: got %f, want 0.9", fake.lastSearchReq.ScoreThreshold)
	}
	if fake.lastSearchReq.TargetURI != "doc://project" {
		t.Errorf("Search TargetURI: got %q, want %q", fake.lastSearchReq.TargetURI, "doc://project")
	}
}

func TestContextSearch_DefaultParams(t *testing.T) {
	fake := &fakeSearcher{
		searchResult: &search.FindResult{Results: []search.Result{}, Total: 0},
	}
	s := buildServer(t, fake)

	result := callTool(t, s, "context_search", map[string]any{
		"query": "defaults",
	})

	if result.IsError {
		t.Fatalf("expected success, got error result")
	}

	if fake.lastSearchReq.Limit != 10 {
		t.Errorf("default Limit: got %d, want 10", fake.lastSearchReq.Limit)
	}
	if fake.lastSearchReq.ScoreThreshold != 0.0 {
		t.Errorf("default ScoreThreshold: got %f, want 0.0", fake.lastSearchReq.ScoreThreshold)
	}
	if fake.lastSearchReq.TargetURI != "" {
		t.Errorf("default TargetURI: got %q, want empty", fake.lastSearchReq.TargetURI)
	}
}

func TestSearcherError_Propagation(t *testing.T) {
	sentinelErr := errors.New("searcher unavailable")
	fake := &fakeSearcher{
		findErr:   sentinelErr,
		searchErr: sentinelErr,
	}
	s := buildServer(t, fake)

	for _, toolName := range []string{"search_memories", "context_search"} {
		t.Run(toolName, func(t *testing.T) {
			result := callTool(t, s, toolName, map[string]any{
				"query": "anything",
			})
			if !result.IsError {
				t.Errorf("%s: expected error result when searcher fails", toolName)
			}
		})
	}
}
