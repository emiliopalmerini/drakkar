package memory_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/epalmerini/drakkar/memory"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/mcptest"
)

// fakeWriter is a fake MemoryWriter that captures inputs and returns a preset result or error.
type fakeWriter struct {
	capturedContent string
	capturedRole    string
	result          *memory.MemoryResult
	err             error
}

func (f *fakeWriter) AddMemory(_ context.Context, content string, role string) (*memory.MemoryResult, error) {
	f.capturedContent = content
	f.capturedRole = role
	return f.result, f.err
}

func buildServer(t *testing.T, fake *fakeWriter) *mcptest.Server {
	t.Helper()
	srv := mcptest.NewUnstartedServer(t)
	memory.Register(srv, fake)
	if err := srv.Start(context.Background()); err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	t.Cleanup(srv.Close)
	return srv
}

func callTool(t *testing.T, s *mcptest.Server, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Name = "add_memory"
	req.Params.Arguments = args
	result, err := s.Client().CallTool(context.Background(), req)
	if err != nil {
		t.Fatalf("CallTool error: %v", err)
	}
	return result
}

func TestAddMemory_WithContentAndRole(t *testing.T) {
	fake := &fakeWriter{
		result: &memory.MemoryResult{URI: "mem://1", Message: "stored"},
	}
	s := buildServer(t, fake)

	result := callTool(t, s, map[string]any{
		"content": "remember this",
		"role":    "assistant",
	})

	if result.IsError {
		t.Fatalf("expected success, got error result")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected non-empty content")
	}
	text, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	var got memory.MemoryResult
	if err := json.Unmarshal([]byte(text.Text), &got); err != nil {
		t.Fatalf("failed to unmarshal result JSON: %v", err)
	}
	if got.URI != "mem://1" || got.Message != "stored" {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestAddMemory_DefaultsRoleToUser(t *testing.T) {
	fake := &fakeWriter{
		result: &memory.MemoryResult{URI: "mem://2", Message: "ok"},
	}
	s := buildServer(t, fake)

	callTool(t, s, map[string]any{
		"content": "hello",
	})

	if fake.capturedRole != "user" {
		t.Errorf("expected role 'user', got %q", fake.capturedRole)
	}
}

func TestAddMemory_EmptyContentReturnsError(t *testing.T) {
	fake := &fakeWriter{
		result: &memory.MemoryResult{URI: "mem://3", Message: "ok"},
	}
	s := buildServer(t, fake)

	result := callTool(t, s, map[string]any{
		"content": "",
	})

	if !result.IsError {
		t.Fatal("expected error result for empty content")
	}
}

func TestAddMemory_PropagatesWriterError(t *testing.T) {
	fake := &fakeWriter{
		err: errors.New("upstream failure"),
	}
	s := buildServer(t, fake)

	result := callTool(t, s, map[string]any{
		"content": "some content",
		"role":    "user",
	})

	if !result.IsError {
		t.Fatal("expected error result when writer returns error")
	}
}

func TestAddMemory_CapturesCorrectValues(t *testing.T) {
	fake := &fakeWriter{
		result: &memory.MemoryResult{URI: "mem://4", Message: "captured"},
	}
	s := buildServer(t, fake)

	callTool(t, s, map[string]any{
		"content": "my important note",
		"role":    "assistant",
	})

	if fake.capturedContent != "my important note" {
		t.Errorf("expected content 'my important note', got %q", fake.capturedContent)
	}
	if fake.capturedRole != "assistant" {
		t.Errorf("expected role 'assistant', got %q", fake.capturedRole)
	}
}
