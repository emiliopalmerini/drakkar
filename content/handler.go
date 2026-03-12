package content

import (
	"context"
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

// Register adds the read_content tool to the MCP server.
func Register(s ToolAdder, reader ContentReader) {
	tool := mcp.NewTool("read_content",
		mcp.WithDescription("Read content at a chosen level of detail: abstract, overview, or full."),
		mcp.WithString("uri", mcp.Required(), mcp.Description("Content URI to read")),
		mcp.WithString("level",
			mcp.Description("Detail level: abstract, overview, or full"),
			mcp.Enum("abstract", "overview", "full"),
		),
	)
	s.AddTool(tool, readContentHandler(reader))
}

func readContentHandler(reader ContentReader) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uri, err := req.RequireString("uri")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("uri is required: %v", err)), nil
		}

		level := req.GetString("level", "overview")

		var text string
		switch level {
		case "abstract":
			text, err = reader.ReadAbstract(ctx, uri)
		case "overview":
			text, err = reader.ReadOverview(ctx, uri)
		case "full":
			text, err = reader.ReadFull(ctx, uri)
		default:
			return mcp.NewToolResultError(fmt.Sprintf("invalid level %q: must be abstract, overview, or full", level)), nil
		}

		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("read content failed: %v", err)), nil
		}

		return mcp.NewToolResultText(text), nil
	}
}
