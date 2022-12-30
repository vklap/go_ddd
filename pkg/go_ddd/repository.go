package go_ddd

import (
	"context"
)

// RollbackCommitter exposes Commit and Rollback methods that can be implemented by repositories.
type RollbackCommitter interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
