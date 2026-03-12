package main

import "github.com/mark3labs/mcp-go/server"

func main() {
	server.ServeStdio(newServer())
}
