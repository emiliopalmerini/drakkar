package mcputil

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ToolAdder is satisfied by *server.MCPServer and *mcptest.Server, allowing
// Register functions to be used in both production wiring and tests without
// coupling to a concrete type.
type ToolAdder interface {
	AddTool(tool mcp.Tool, handler server.ToolHandlerFunc)
}
