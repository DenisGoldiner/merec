package merec

import (
	"context"
	"fmt"
	"time"
)

type timeoutOption[In, Out any] struct {
	timeout time.Duration
}

// NewTimeoutOption is a constructor for the timeoutOption.
func NewTimeoutOption[In, Out any](timeout time.Duration) CallOption[In, Out] {
	return timeoutOption[In, Out]{timeout: timeout}
}

// WithOption implements the CallOption interface for the timeoutOption.
func (to timeoutOption[In, Out]) WithOption(next Call[In, Out]) Call[In, Out] {
	return func(ctx context.Context, in In) (Out, error) {
		ctx, ctxCsl := context.WithTimeout(ctx, to.timeout)
		defer ctxCsl()

		return next(ctx, in)
	}
}

type failFastOption[In, Out any] struct {
	mistakesLimit int
}

// NewFailFastOptionOption is a constructor for the failFastOption.
func NewFailFastOptionOption[In, Out any](mistakesLimit int) CallOption[In, Out] {
	return failFastOption[In, Out]{mistakesLimit: mistakesLimit}
}

// WithOption implements the CallOption interface for the failFastOption.
func (failFastOption[In, Out]) WithOption(next Call[In, Out]) Call[In, Out] {
	return func(ctx context.Context, in In) (Out, error) {
		out, err := next(ctx, in)
		if err != nil {
			return *new(Out), fmt.Errorf("%w: %w", ErrMustStop, err)
		}

		return out, nil
	}
}
