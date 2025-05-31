package merec

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// RunWorkerPool starts the pool of goroutine workers to consume from the input channel and execute
// the call functions with inputs. Returns the channel to be listened to, to get the Result values.
// Executions are independent, and it doesn't stop processing inputs if call fails.
func RunWorkerPool[In, Out any](
	ctx context.Context,
	inCh <-chan In,
	call Call[In, Out],
	poolSize int,
	options ...CallOption[In, Out],
) (<-chan Result[Out], error) {
	next := call

	for _, o := range options {
		next = o.WithOption(next)
	}

	return runWorkerPool(ctx, inCh, next, poolSize)
}

func runWorkerPool[In, Out any](
	ctx context.Context,
	inCh <-chan In,
	call Call[In, Out],
	poolSize int,
) (<-chan Result[Out], error) {
	if err := validateRunFromChanInputs(ctx, inCh, call); err != nil {
		return nil, err
	}

	ctx, ctxCsl := context.WithCancel(ctx)

	// TODO: define buf size
	resChanPool := SpawnResChanPool[Result[Out]](poolSize, 0)

	worker := func(resCh chan Result[Out]) {
		defer close(resCh)

		for in := range inCh {
			res, err := call(ctx, in)
			if errors.Is(err, ErrMustStop) {
				resCh <- ErrorResult[Out](fmt.Errorf("%w: %w", ErrBusinessLogic, err))

				ctxCsl()

				return
			}

			if err != nil {
				resCh <- ErrorResult[Out](fmt.Errorf("%w: %w", ErrBusinessLogic, err))
				continue
			}

			resCh <- ValueResult[Out](res)
		}
	}

	for i := 0; i < poolSize; i++ {
		go worker(resChanPool[i])
	}

	return MergeChanPool(resChanPool), nil
}

// SpawnResChanPool creates as many channels as it is needed.
// Implements concurrency pattern FanOut.
func SpawnResChanPool[T any](poolSize int, bufSize int) []chan T {
	chanPool := make([]chan T, poolSize)

	for i := range chanPool {
		chanPool[i] = make(chan T, bufSize)
	}

	return chanPool
}

// MergeChanPool combines the output from lit of channels into a single one.
// Implements concurrency pattern FanIn.
func MergeChanPool[T any](resChanPool []chan T) chan T {
	mergeCh := make(chan T, len(resChanPool))

	var wg sync.WaitGroup

	wg.Add(len(resChanPool))

	for _, rCh := range resChanPool {
		go func(resCh chan T) {
			for out := range resCh {
				mergeCh <- out
			}

			wg.Done()
		}(rCh)
	}

	go func() {
		wg.Wait()
		close(mergeCh)
	}()

	return mergeCh
}

// MergeSignalChanPool combines the output from lit of channels into a single one.
// Implements concurrency pattern FanIn. In addition, it takes the value from the chan only
// if there is a corresponding signal.
func MergeSignalChanPool[T any](inChanPool []chan T, signalChanPool []chan struct{}) (<-chan struct{}, chan T) {
	done := make(chan struct{})
	mergeCh := make(chan T, len(inChanPool))

	var wg sync.WaitGroup

	wg.Add(len(inChanPool))

	for i, inCh := range inChanPool {
		go withSignal(&wg, inCh, signalChanPool[i], mergeCh)
	}

	go func() {
		wg.Wait()
		close(mergeCh)
		done <- struct{}{}
	}()

	return done, mergeCh
}

func withSignal[T any](
	wg *sync.WaitGroup,
	inCh <-chan T,
	signalCh <-chan struct{},
	mergeCh chan<- T,
) {
	for out := range inCh {
		<-signalCh
		mergeCh <- out
	}

	wg.Done()
}

// TrySend tries to send the value into the channel.
// If channel is blocked we do nothing and return.
func TrySend[T any](ch chan T, v T) {
	select {
	case ch <- v:
	default:
	}
}

// TryReedSignal tires to read the signal from the channel.
// If there are no values in the channel we do nothing and return.
func TryReedSignal(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}
