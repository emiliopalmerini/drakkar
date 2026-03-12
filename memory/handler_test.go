package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/epalmerini/drakkar/memory"
	mcptestutil "github.com/epalmerini/drakkar/internal/mcptest"
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
	return mcptestutil.StartServer(t, func(srv *mcptest.Server) {
		memory.Register(srv, fake)
	})
}

func TestAddMemory_WithContentAndRole(t *testing.T) {
	fake := &fakeWriter{
		result: &memory.MemoryResult{URI: "mem://1", Message: "stored"},
	}
	s := buildServer(t, fake)

	result := mcptestutil.CallTool(t, s, "add_memory", map[string]any{
		"content": "remember this",
		"role":    "assistant",
	})

	if result.IsError {
		t.Fatalf("expected success, got error result")
	}

	var got memory.MemoryResult
	mcptestutil.DecodeText(t, result, &got)

	if got.URI != "mem://1" || got.Message != "stored" {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestAddMemory_DefaultsRoleToUser(t *testing.T) {
	fake := &fakeWriter{
		result: &memory.MemoryResult{URI: "mem://2", Message: "ok"},
	}
	s := buildServer(t, fake)

	mcptestutil.CallTool(t, s, "add_memory", map[string]any{
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

	result := mcptestutil.CallTool(t, s, "add_memory", map[string]any{
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

	result := mcptestutil.CallTool(t, s, "add_memory", map[string]any{
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

	mcptestutil.CallTool(t, s, "add_memory", map[string]any{
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
