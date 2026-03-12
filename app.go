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

func newServer() *server.MCPServer {
	url := os.Getenv("OPENVIKING_URL")
	if url == "" {
		url = "http://localhost:1933"
	}

	client := openviking.NewClient(url)

	s := server.NewMCPServer("drakkar", "0.1.0", server.WithToolCapabilities(true))
	search.Register(s, client)
	content.Register(s, client)
	browse.Register(s, client)
	memory.Register(s, client)

	return s
}
