package memory

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ToolAdder is satisfied by *server.MCPServer and *mcptest.Server, allowing
// Register to be used in both production wiring and tests without coupling to
// a concrete type.
type ToolAdder interface {
	AddTool(tool mcp.Tool, handler server.ToolHandlerFunc)
}

// Register wires the add_memory tool into the MCP server.
func Register(s ToolAdder, writer MemoryWriter) {
	tool := mcp.NewTool("add_memory",
		mcp.WithDescription("Store a memory entry with an optional role"),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("The content of the memory to store"),
		),
		mcp.WithString("role",
			mcp.Description("The role associated with the memory (e.g. 'user' or 'assistant')"),
			mcp.DefaultString("user"),
		),
	)
	s.AddTool(tool, addMemoryHandler(writer))
}

func addMemoryHandler(writer MemoryWriter) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		content, err := req.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("content is required: %v", err)), nil
		}
		if content == "" {
			return mcp.NewToolResultError("content must not be empty"), nil
		}

		role := req.GetString("role", "user")

		result, err := writer.AddMemory(ctx, content, role)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("add memory failed: %v", err)), nil
		}

		data, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("marshal result failed: %v", err)), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}
