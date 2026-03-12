package browse_test

import (
	"context"
	"errors"
	"testing"

	"github.com/epalmerini/drakkar/browse"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/mcptest"
)

// fakeBrowser is a test double that records calls and returns canned responses.
type fakeBrowser struct {
	calledMethod string
	calledURI    string
	result       string
	err          error
}

func (f *fakeBrowser) List(_ context.Context, uri string) (string, error) {
	f.calledMethod = "List"
	f.calledURI = uri
	return f.result, f.err
}

func (f *fakeBrowser) Tree(_ context.Context, uri string) (string, error) {
	f.calledMethod = "Tree"
	f.calledURI = uri
	return f.result, f.err
}

func (f *fakeBrowser) Stat(_ context.Context, uri string) (string, error) {
	f.calledMethod = "Stat"
	f.calledURI = uri
	return f.result, f.err
}

func buildServer(t *testing.T, fake *fakeBrowser) *mcptest.Server {
	t.Helper()
	srv := mcptest.NewUnstartedServer(t)
	browse.Register(srv, fake)
	if err := srv.Start(context.Background()); err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	t.Cleanup(srv.Close)
	return srv
}

func callTool(t *testing.T, s *mcptest.Server, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Name = "browse"
	req.Params.Arguments = args
	result, err := s.Client().CallTool(context.Background(), req)
	if err != nil {
		t.Fatalf("CallTool error: %v", err)
	}
	return result
}

func textContent(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("result has no content")
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	return tc.Text
}

func TestBrowse_ModeLS_CallsList(t *testing.T) {
	fb := &fakeBrowser{result: "file1.txt\nfile2.txt"}
	s := buildServer(t, fb)

	result := callTool(t, s, map[string]any{
		"uri":  "file:///tmp",
		"mode": "ls",
	})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", textContent(t, result))
	}
	if fb.calledMethod != "List" {
		t.Errorf("expected List to be called, got %q", fb.calledMethod)
	}
	if fb.calledURI != "file:///tmp" {
		t.Errorf("expected URI %q, got %q", "file:///tmp", fb.calledURI)
	}
	if got := textContent(t, result); got != "file1.txt\nfile2.txt" {
		t.Errorf("unexpected content: %q", got)
	}
}

func TestBrowse_ModeTree_CallsTree(t *testing.T) {
	fb := &fakeBrowser{result: ".\n└── file1.txt"}
	s := buildServer(t, fb)

	result := callTool(t, s, map[string]any{
		"uri":  "file:///tmp",
		"mode": "tree",
	})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", textContent(t, result))
	}
	if fb.calledMethod != "Tree" {
		t.Errorf("expected Tree to be called, got %q", fb.calledMethod)
	}
	if got := textContent(t, result); got != ".\n└── file1.txt" {
		t.Errorf("unexpected content: %q", got)
	}
}

func TestBrowse_ModeStat_CallsStat(t *testing.T) {
	fb := &fakeBrowser{result: "size: 1024"}
	s := buildServer(t, fb)

	result := callTool(t, s, map[string]any{
		"uri":  "file:///tmp/file.txt",
		"mode": "stat",
	})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", textContent(t, result))
	}
	if fb.calledMethod != "Stat" {
		t.Errorf("expected Stat to be called, got %q", fb.calledMethod)
	}
	if fb.calledURI != "file:///tmp/file.txt" {
		t.Errorf("expected URI %q, got %q", "file:///tmp/file.txt", fb.calledURI)
	}
	if got := textContent(t, result); got != "size: 1024" {
		t.Errorf("unexpected content: %q", got)
	}
}

func TestBrowse_NoMode_DefaultsToLS(t *testing.T) {
	fb := &fakeBrowser{result: "dir/"}
	s := buildServer(t, fb)

	result := callTool(t, s, map[string]any{
		"uri": "file:///home",
	})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", textContent(t, result))
	}
	if fb.calledMethod != "List" {
		t.Errorf("expected List to be called by default, got %q", fb.calledMethod)
	}
}

func TestBrowse_MissingURI_ReturnsError(t *testing.T) {
	fb := &fakeBrowser{}
	s := buildServer(t, fb)

	result := callTool(t, s, map[string]any{
		"mode": "ls",
	})

	if !result.IsError {
		t.Fatal("expected error result when uri is missing")
	}
	if fb.calledMethod != "" {
		t.Errorf("expected no browser method to be called, got %q", fb.calledMethod)
	}
}

func TestBrowse_InvalidMode_ReturnsError(t *testing.T) {
	fb := &fakeBrowser{}
	s := buildServer(t, fb)

	result := callTool(t, s, map[string]any{
		"uri":  "file:///tmp",
		"mode": "walk",
	})

	if !result.IsError {
		t.Fatal("expected error result for invalid mode")
	}
	if fb.calledMethod != "" {
		t.Errorf("expected no browser method to be called, got %q", fb.calledMethod)
	}
}

func TestBrowse_BrowserError_PropagatesAsError(t *testing.T) {
	fb := &fakeBrowser{err: errors.New("connection refused")}
	s := buildServer(t, fb)

	result := callTool(t, s, map[string]any{
		"uri":  "file:///tmp",
		"mode": "ls",
	})

	if !result.IsError {
		t.Fatal("expected error result when browser returns error")
	}
	content := textContent(t, result)
	if content == "" {
		t.Error("expected non-empty error message")
	}
}
