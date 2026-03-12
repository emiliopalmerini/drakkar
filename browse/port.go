package browse

import "context"

// Browser is the port interface for directory/file browsing operations.
// Implementations are responsible for resolving URIs and returning
// human-readable output strings.
type Browser interface {
	// List returns a flat listing of the entries at uri (analogous to ls).
	List(ctx context.Context, uri string) (string, error)

	// Tree returns a recursive tree view of the entries at uri.
	Tree(ctx context.Context, uri string) (string, error)

	// Stat returns metadata about the entry at uri.
	Stat(ctx context.Context, uri string) (string, error)
}
