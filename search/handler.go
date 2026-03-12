package search

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/epalmerini/drakkar/internal/mcputil"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// searchFunc is the signature shared by Searcher.Find and Searcher.Search.
type searchFunc func(ctx context.Context, req Request) (*FindResult, error)

// Register wires the search_memories and context_search MCP tools onto s.
func Register(s mcputil.ToolAdder, searcher Searcher) {
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
	s.AddTool(findTool, makeSearchHandler(searcher.Find))

	searchTool := mcp.NewTool("context_search",
		append([]mcp.ToolOption{
			mcp.WithDescription("Search current context by semantic similarity"),
		}, commonParams...)...,
	)
	s.AddTool(searchTool, makeSearchHandler(searcher.Search))
}

func makeSearchHandler(fn searchFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("query is required: %v", err)), nil
		}
		if query == "" {
			return mcp.NewToolResultError("query must not be empty"), nil
		}

		searchReq := Request{
			Query:          query,
			Limit:          req.GetInt("limit", 10),
			ScoreThreshold: req.GetFloat("score_threshold", 0.0),
			TargetURI:      req.GetString("target_uri", ""),
		}

		result, err := fn(ctx, searchReq)
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
