package browse_test

import (
	"context"
	"errors"
	"testing"

	"github.com/epalmerini/drakkar/browse"
	mcptestutil "github.com/epalmerini/drakkar/internal/mcptest"
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
	return mcptestutil.StartServer(t, func(srv *mcptest.Server) {
		browse.Register(srv, fake)
	})
}

func TestBrowse_ModeLS_CallsList(t *testing.T) {
	fb := &fakeBrowser{result: "file1.txt\nfile2.txt"}
	s := buildServer(t, fb)

	result := mcptestutil.CallTool(t, s, "browse", map[string]any{
		"uri":  "file:///tmp",
		"mode": "ls",
	})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", mcptestutil.TextContent(t, result))
	}
	if fb.calledMethod != "List" {
		t.Errorf("expected List to be called, got %q", fb.calledMethod)
	}
	if fb.calledURI != "file:///tmp" {
		t.Errorf("expected URI %q, got %q", "file:///tmp", fb.calledURI)
	}
	if got := mcptestutil.TextContent(t, result); got != "file1.txt\nfile2.txt" {
		t.Errorf("unexpected content: %q", got)
	}
}

func TestBrowse_ModeTree_CallsTree(t *testing.T) {
	fb := &fakeBrowser{result: ".\n└── file1.txt"}
	s := buildServer(t, fb)

	result := mcptestutil.CallTool(t, s, "browse", map[string]any{
		"uri":  "file:///tmp",
		"mode": "tree",
	})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", mcptestutil.TextContent(t, result))
	}
	if fb.calledMethod != "Tree" {
		t.Errorf("expected Tree to be called, got %q", fb.calledMethod)
	}
	if got := mcptestutil.TextContent(t, result); got != ".\n└── file1.txt" {
		t.Errorf("unexpected content: %q", got)
	}
}

func TestBrowse_ModeStat_CallsStat(t *testing.T) {
	fb := &fakeBrowser{result: "size: 1024"}
	s := buildServer(t, fb)

	result := mcptestutil.CallTool(t, s, "browse", map[string]any{
		"uri":  "file:///tmp/file.txt",
		"mode": "stat",
	})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", mcptestutil.TextContent(t, result))
	}
	if fb.calledMethod != "Stat" {
		t.Errorf("expected Stat to be called, got %q", fb.calledMethod)
	}
	if fb.calledURI != "file:///tmp/file.txt" {
		t.Errorf("expected URI %q, got %q", "file:///tmp/file.txt", fb.calledURI)
	}
	if got := mcptestutil.TextContent(t, result); got != "size: 1024" {
		t.Errorf("unexpected content: %q", got)
	}
}

func TestBrowse_NoMode_DefaultsToLS(t *testing.T) {
	fb := &fakeBrowser{result: "dir/"}
	s := buildServer(t, fb)

	result := mcptestutil.CallTool(t, s, "browse", map[string]any{
		"uri": "file:///home",
	})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", mcptestutil.TextContent(t, result))
	}
	if fb.calledMethod != "List" {
		t.Errorf("expected List to be called by default, got %q", fb.calledMethod)
	}
}

func TestBrowse_MissingURI_ReturnsError(t *testing.T) {
	fb := &fakeBrowser{}
	s := buildServer(t, fb)

	result := mcptestutil.CallTool(t, s, "browse", map[string]any{
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

	result := mcptestutil.CallTool(t, s, "browse", map[string]any{
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

	result := mcptestutil.CallTool(t, s, "browse", map[string]any{
		"uri":  "file:///tmp",
		"mode": "ls",
	})

	if !result.IsError {
		t.Fatal("expected error result when browser returns error")
	}
	content := mcptestutil.TextContent(t, result)
	if content == "" {
		t.Error("expected non-empty error message")
	}
}
