package merec

import (
	"context"
	"fmt"
)

// RunFromInput executes the call function in a separate goroutine with the specified input.
// Returns the channel to be listened to, to get the Result value.
func RunFromInput[In, Out any](
	ctx context.Context,
	in In,
	call Call[In, Out],
	options ...CallOption[In, Out],
) (<-chan Result[Out], error) {
	next := call

	for _, o := range options {
		next = o.WithOption(next)
	}

	return runFromInput(ctx, in, next)
}

func runFromInput[In, Out any](ctx context.Context, in In, call Call[In, Out]) (<-chan Result[Out], error) {
	if err := validateRunFromInputInputs(ctx, call); err != nil {
		return nil, err
	}

	resCh := make(chan Result[Out], 1)

	go func() {
		defer close(resCh)

		res, err := call(ctx, in)
		if err != nil {
			resCh <- ErrorResult[Out](fmt.Errorf("%w: %w", ErrBusinessLogic, err))

			return
		}

		resCh <- ValueResult[Out](res)
	}()

	return resCh, nil
}

func validateRunFromInputInputs[In, Out any](ctx context.Context, call Call[In, Out]) error {
	if ctx == nil {
		return ErrNilContext
	}

	if call == nil {
		return ErrNilCallFunc
	}

	return nil
}
