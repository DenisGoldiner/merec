package merec

import (
	"context"
	"errors"
	"fmt"
)

// RunFromChan starts a separate goroutine to consume from the input channel and execute the call function with it.
// Returns the channel to be listened to, to get the Result values.
// Executions are independent, and it doesn't stop processing inputs if call fails.
func RunFromChan[In, Out any](
	ctx context.Context,
	inCh <-chan In,
	call Call[In, Out],
	options ...CallOption[In, Out],
) (<-chan Result[Out], error) {
	next := call

	for _, o := range options {
		next = o.WithOption(next)
	}

	return runFromChan(ctx, inCh, next)
}

func runFromChan[In, Out any](ctx context.Context, inCh <-chan In, call Call[In, Out]) (<-chan Result[Out], error) {
	if err := validateRunFromChanInputs(ctx, inCh, call); err != nil {
		return nil, err
	}

	resCh := make(chan Result[Out], cap(inCh))

	go func() {
		defer close(resCh)

		for in := range inCh {
			res, err := call(ctx, in)
			if errors.Is(err, ErrMustStop) {
				resCh <- ErrorResult[Out](fmt.Errorf("%w: %w", ErrBusinessLogic, err))
				return
			}

			if err != nil {
				resCh <- ErrorResult[Out](fmt.Errorf("%w: %w", ErrBusinessLogic, err))
				continue
			}

			resCh <- ValueResult[Out](res)
		}
	}()

	return resCh, nil
}

func validateRunFromChanInputs[In, Out any](ctx context.Context, inCh <-chan In, call Call[In, Out]) error {
	if ctx == nil {
		return ErrNilContext
	}

	if inCh == nil {
		return ErrNilInChan
	}

	if call == nil {
		return ErrNilCallFunc
	}

	return nil
}
