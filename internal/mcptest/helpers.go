package mcptest

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/mcptest"
)

// StartServer starts an mcptest server with the given registration function
// and returns it ready to use. It registers cleanup automatically.
func StartServer(t *testing.T, register func(srv *mcptest.Server)) *mcptest.Server {
	t.Helper()
	srv := mcptest.NewUnstartedServer(t)
	register(srv)
	if err := srv.Start(context.Background()); err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	t.Cleanup(srv.Close)
	return srv
}

// CallTool invokes a named tool with an arguments map.
func CallTool(t *testing.T, s *mcptest.Server, toolName string, args map[string]any) *mcp.CallToolResult {
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

// TextContent extracts the text from the first content item in a result.
func TextContent(t *testing.T, result *mcp.CallToolResult) string {
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

// DecodeText unmarshals the first text content from a result into dst.
func DecodeText(t *testing.T, result *mcp.CallToolResult, dst any) {
	t.Helper()
	text := TextContent(t, result)
	if err := json.Unmarshal([]byte(text), dst); err != nil {
		t.Fatalf("unmarshal result text: %v", err)
	}
}
