package content_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/epalmerini/drakkar/content"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/mcptest"
)

// fakeReader is a test double that tracks which method was called and with what URI.
type fakeReader struct {
	calledMethod string
	calledURI    string
	result       string
	err          error
}

func (f *fakeReader) ReadAbstract(_ context.Context, uri string) (string, error) {
	f.calledMethod = "abstract"
	f.calledURI = uri
	return f.result, f.err
}

func (f *fakeReader) ReadOverview(_ context.Context, uri string) (string, error) {
	f.calledMethod = "overview"
	f.calledURI = uri
	return f.result, f.err
}

func (f *fakeReader) ReadFull(_ context.Context, uri string) (string, error) {
	f.calledMethod = "full"
	f.calledURI = uri
	return f.result, f.err
}

func buildServer(t *testing.T, fake *fakeReader) *mcptest.Server {
	t.Helper()
	srv := mcptest.NewUnstartedServer(t)
	content.Register(srv, fake)
	if err := srv.Start(context.Background()); err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	t.Cleanup(srv.Close)
	return srv
}

func callTool(t *testing.T, s *mcptest.Server, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Name = "read_content"
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

func TestReadContent_AbstractLevel(t *testing.T) {
	fake := &fakeReader{result: "abstract text"}
	s := buildServer(t, fake)

	result := callTool(t, s, map[string]any{
		"uri":   "mem://some/path",
		"level": "abstract",
	})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", textContent(t, result))
	}
	if fake.calledMethod != "abstract" {
		t.Errorf("expected ReadAbstract to be called, got %q", fake.calledMethod)
	}
	if fake.calledURI != "mem://some/path" {
		t.Errorf("expected URI %q, got %q", "mem://some/path", fake.calledURI)
	}
	if got := textContent(t, result); got != "abstract text" {
		t.Errorf("expected %q, got %q", "abstract text", got)
	}
}

func TestReadContent_OverviewLevel(t *testing.T) {
	fake := &fakeReader{result: "overview text"}
	s := buildServer(t, fake)

	result := callTool(t, s, map[string]any{
		"uri":   "mem://another/path",
		"level": "overview",
	})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", textContent(t, result))
	}
	if fake.calledMethod != "overview" {
		t.Errorf("expected ReadOverview to be called, got %q", fake.calledMethod)
	}
	if got := textContent(t, result); got != "overview text" {
		t.Errorf("expected %q, got %q", "overview text", got)
	}
}

func TestReadContent_FullLevel(t *testing.T) {
	fake := &fakeReader{result: "full text"}
	s := buildServer(t, fake)

	result := callTool(t, s, map[string]any{
		"uri":   "mem://full/path",
		"level": "full",
	})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", textContent(t, result))
	}
	if fake.calledMethod != "full" {
		t.Errorf("expected ReadFull to be called, got %q", fake.calledMethod)
	}
	if got := textContent(t, result); got != "full text" {
		t.Errorf("expected %q, got %q", "full text", got)
	}
}

func TestReadContent_DefaultLevelIsOverview(t *testing.T) {
	fake := &fakeReader{result: "default overview text"}
	s := buildServer(t, fake)

	result := callTool(t, s, map[string]any{
		"uri": "mem://default/path",
	})

	if result.IsError {
		t.Fatalf("expected success, got error: %s", textContent(t, result))
	}
	if fake.calledMethod != "overview" {
		t.Errorf("expected ReadOverview as default, got %q", fake.calledMethod)
	}
}

func TestReadContent_MissingURI(t *testing.T) {
	fake := &fakeReader{}
	s := buildServer(t, fake)

	result := callTool(t, s, map[string]any{
		"level": "overview",
	})

	if !result.IsError {
		t.Fatal("expected error result when uri is missing")
	}
	if fake.calledMethod != "" {
		t.Errorf("expected no reader method to be called, got %q", fake.calledMethod)
	}
}

func TestReadContent_InvalidLevel(t *testing.T) {
	fake := &fakeReader{}
	s := buildServer(t, fake)

	result := callTool(t, s, map[string]any{
		"uri":   "mem://some/path",
		"level": "nonexistent",
	})

	if !result.IsError {
		t.Fatal("expected error result for invalid level")
	}
	if fake.calledMethod != "" {
		t.Errorf("expected no reader method to be called, got %q", fake.calledMethod)
	}
}

func TestReadContent_ErrorPropagation(t *testing.T) {
	fake := &fakeReader{err: fmt.Errorf("storage unavailable")}
	s := buildServer(t, fake)

	result := callTool(t, s, map[string]any{
		"uri":   "mem://broken/path",
		"level": "full",
	})

	if !result.IsError {
		t.Fatal("expected error result when reader returns error")
	}
	if fake.calledMethod != "full" {
		t.Errorf("expected ReadFull to be called, got %q", fake.calledMethod)
	}
}
