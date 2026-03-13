package memory

import "context"

// MemoryResult is returned by a successful AddMemory call.
type MemoryResult struct {
	URI     string `json:"uri"`
	Message string `json:"message"`
}

// CommitResult is returned by a successful CommitSession call.
type CommitResult struct {
	SessionID         string `json:"session_id"`
	Status            string `json:"status"`
	MemoriesExtracted int    `json:"memories_extracted"`
	Archived          bool   `json:"archived"`
}

// MemoryWriter is the port for persisting memory entries and committing sessions.
type MemoryWriter interface {
	AddMemory(ctx context.Context, content string, role string) (*MemoryResult, error)
	CommitSession(ctx context.Context) (*CommitResult, error)
}
