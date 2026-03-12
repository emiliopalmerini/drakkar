package browse

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

// Register adds the browse tool to the MCP server.
func Register(s ToolAdder, browser Browser) {
	tool := mcp.NewTool("browse",
		mcp.WithDescription("Browse files and directories at a given URI"),
		mcp.WithString("uri",
			mcp.Required(),
			mcp.Description("URI of the resource to browse"),
		),
		mcp.WithString("mode",
			mcp.Description("Browse mode: ls (default), tree, or stat"),
			mcp.Enum("ls", "tree", "stat"),
		),
	)
	s.AddTool(tool, browseHandler(browser))
}

func browseHandler(browser Browser) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uri, err := req.RequireString("uri")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("uri is required: %v", err)), nil
		}

		mode := req.GetString("mode", "ls")

		var result string
		switch mode {
		case "ls":
			result, err = browser.List(ctx, uri)
		case "tree":
			result, err = browser.Tree(ctx, uri)
		case "stat":
			result, err = browser.Stat(ctx, uri)
		default:
			return mcp.NewToolResultError(fmt.Sprintf("invalid mode %q: must be ls, tree, or stat", mode)), nil
		}

		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("browse failed: %v", err)), nil
		}

		return mcp.NewToolResultText(result), nil
	}
}
