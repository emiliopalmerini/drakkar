package search

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

// Register wires the search_memories and context_search MCP tools onto s.
func Register(s ToolAdder, searcher Searcher) {
	commonParams := []mcp.ToolOption{
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query text"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of results to return (default 10)"),
		),
		mcp.WithNumber("score_threshold",
			mcp.Description("Minimum relevance score threshold (default 0.0)"),
		),
		mcp.WithString("target_uri",
			mcp.Description("Restrict search to a specific target URI"),
		),
	}

	findTool := mcp.NewTool("search_memories",
		append([]mcp.ToolOption{
			mcp.WithDescription("Search stored memories by semantic similarity"),
		}, commonParams...)...,
	)
	s.AddTool(findTool, findHandler(searcher))

	searchTool := mcp.NewTool("context_search",
		append([]mcp.ToolOption{
			mcp.WithDescription("Search current context by semantic similarity"),
		}, commonParams...)...,
	)
	s.AddTool(searchTool, searchHandler(searcher))
}

func findHandler(searcher Searcher) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("query is required: %v", err)), nil
		}
		if query == "" {
			return mcp.NewToolResultError("query must not be empty"), nil
		}

		findReq := FindRequest{
			Query:          query,
			Limit:          req.GetInt("limit", 10),
			ScoreThreshold: req.GetFloat("score_threshold", 0.0),
			TargetURI:      req.GetString("target_uri", ""),
		}

		result, err := searcher.Find(ctx, findReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("find failed: %v", err)), nil
		}

		data, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("marshal result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}

func searchHandler(searcher Searcher) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("query is required: %v", err)), nil
		}
		if query == "" {
			return mcp.NewToolResultError("query must not be empty"), nil
		}

		searchReq := SearchRequest{
			Query:          query,
			Limit:          req.GetInt("limit", 10),
			ScoreThreshold: req.GetFloat("score_threshold", 0.0),
			TargetURI:      req.GetString("target_uri", ""),
		}

		result, err := searcher.Search(ctx, searchReq)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
		}

		data, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("marshal result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}
