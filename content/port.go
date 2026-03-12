package content

import "context"

// ContentReader is the port that delivers document content at varying levels of detail.
type ContentReader interface {
	ReadAbstract(ctx context.Context, uri string) (string, error)
	ReadOverview(ctx context.Context, uri string) (string, error)
	ReadFull(ctx context.Context, uri string) (string, error)
}
