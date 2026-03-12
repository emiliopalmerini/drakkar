package memory

import "context"

// MemoryResult is returned by a successful AddMemory call.
type MemoryResult struct {
	URI     string `json:"uri"`
	Message string `json:"message"`
}

// MemoryWriter is the port for persisting a memory entry.
type MemoryWriter interface {
	AddMemory(ctx context.Context, content string, role string) (*MemoryResult, error)
}
