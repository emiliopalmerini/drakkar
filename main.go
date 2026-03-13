package main

import (
	"context"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	a := newApp()

	if err := server.ServeStdio(a.server); err != nil {
		log.Fatal(err)
	}

	// Auto-commit the session on shutdown to promote messages to persistent memory.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := a.client.CommitSession(ctx); err != nil {
		log.Printf("commit session: %v", err)
	}
}
