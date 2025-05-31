package merec_test

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/DenisGoldiner/merec"
)

const (
	workLoad = 5
)

var (
	errCtxCancel   = errors.New("root context was canceled")
	errCtxDeadline = errors.New("root context's deadline passed")
)

func stabCall(duration time.Duration) merec.Call[string, int] {
	return func(ctx context.Context, in string) (int, error) {
		timer := time.NewTimer(duration)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return 0, errCtxDeadline
			}

			return 0, errCtxCancel
		case <-timer.C:
			return strconv.Atoi(in)
		}
	}
}

func givenCh(bufSize int) chan string {
	ch := make(chan string, bufSize)

	go func() {
		for i := 0; i < workLoad; i++ {
			ch <- strconv.Itoa(i)
		}
		close(ch)
	}()

	return ch
}

func expectedResults() []merec.Result[int] {
	values := make([]merec.Result[int], workLoad)
	for i := range values {
		values[i] = merec.ValueResult(i)
	}

	return values
}
