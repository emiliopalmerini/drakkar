package main

import (
	"os"

	"github.com/epalmerini/drakkar/browse"
	"github.com/epalmerini/drakkar/content"
	"github.com/epalmerini/drakkar/memory"
	"github.com/epalmerini/drakkar/openviking"
	"github.com/epalmerini/drakkar/search"
	"github.com/mark3labs/mcp-go/server"
)

// app holds the MCP server and the OpenViking client for lifecycle management.
type app struct {
	server *server.MCPServer
	client *openviking.Client
}

func newApp() *app {
	url := os.Getenv("OPENVIKING_URL")
	if url == "" {
		url = "http://localhost:1933"
	}
	apiKey := os.Getenv("OPENVIKING_API_KEY")

	client := openviking.NewClient(url, apiKey)

	s := server.NewMCPServer("drakkar", "0.1.0", server.WithToolCapabilities(true))
	search.Register(s, client)
	content.Register(s, client)
	browse.Register(s, client)
	memory.Register(s, client)

	return &app{server: s, client: client}
}
