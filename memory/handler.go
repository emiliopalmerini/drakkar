package memory

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/epalmerini/drakkar/internal/mcputil"
	"github.com/mark3labs/mcp-go/mcp"
)

// Register wires the add_memory tool into the MCP server.
func Register(s mcputil.ToolAdder, writer MemoryWriter) {
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

func addMemoryHandler(writer MemoryWriter) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
