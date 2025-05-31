package merec

import (
	"context"
	"errors"
)

// CheckContext checks if the context is done without blocking the execution.
func CheckContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return ErrCtxDeadline
		}

		return ErrCtxCancel

	default:
		return nil
	}
}
