package merec_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/DenisGoldiner/merec"
)

func TestRunFromChan(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]struct {
		givenIn chan string
	}{
		"buf_size_0": {
			givenIn: givenCh(0),
		},
		"buf_size_workLoad": {
			givenIn: givenCh(workLoad),
		},
	}

	for tcName, tc := range testCases {
		tc := tc

		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			resCh, err := merec.RunFromChan(ctx, tc.givenIn, stabCall(10*time.Microsecond))
			require.NoError(t, err)

			results := make([]merec.Result[int], 0, workLoad)

			for r := range resCh {
				results = append(results, r)
			}

			require.ElementsMatch(t, results, expectedResults())
		})
	}
}

func TestRunFromChan_Errors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]struct {
		givenIn     func() chan string
		expValueRes []merec.Result[int]
		expErrRes   []merec.Result[int]
	}{
		"single_business_error": {
			givenIn: func() chan string {
				ch := make(chan string)

				go func() {
					ch <- "qwerty"
					close(ch)
				}()

				return ch
			},
			expErrRes: []merec.Result[int]{
				merec.ErrorResult[int](merec.ErrBusinessLogic),
			},
		},
		"success_and_business_error": {
			givenIn: func() chan string {
				ch := make(chan string)

				go func() {
					for i := 0; i < workLoad; i++ {
						if i == 3 {
							ch <- "qwerty"
							continue
						}
						ch <- strconv.Itoa(i)
					}
					close(ch)
				}()

				return ch
			},
			expErrRes: []merec.Result[int]{
				merec.ErrorResult[int](merec.ErrBusinessLogic),
			},
			expValueRes: []merec.Result[int]{
				merec.ValueResult(0),
				merec.ValueResult(1),
				merec.ValueResult(2),
				merec.ValueResult(4),
			},
		},
	}

	for tcName, tc := range testCases {
		tc := tc

		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			resCh, err := merec.RunFromChan(ctx, tc.givenIn(), stabCall(10*time.Millisecond))
			require.NoError(t, err)

			var valResults []merec.Result[int]
			var errResults []merec.Result[int]

			for r := range resCh {
				if err := r.Err(); err != nil {
					errResults = append(errResults, r)
					continue
				}

				valResults = append(valResults, r)
			}

			var counter int
			require.ElementsMatch(t, valResults, tc.expValueRes)
			for _, expErr := range tc.expErrRes {
				for _, actErr := range errResults {
					if errors.Is(actErr.Err(), expErr.Err()) {
						counter++
						break
					}
				}
			}
			require.Len(t, errResults, counter)
		})
	}
}

func TestRunFromChan_ValidationFail(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		givenCtx  context.Context
		givenIn   func() chan string
		givenCall merec.Call[string, int]
		expErr    error
	}{
		"nil_ctx": {
			givenIn: func() chan string {
				ch := make(chan string)

				go func() {
					for i := 0; i < workLoad; i++ {
						ch <- strconv.Itoa(i)
					}
				}()

				return ch
			},
			givenCall: stabCall(time.Second),
			expErr:    merec.ErrNilContext,
		},
		"nil_in_chan": {
			givenCtx:  context.Background(),
			givenIn:   func() chan string { return nil },
			givenCall: stabCall(time.Second),
			expErr:    merec.ErrNilInChan,
		},
		"nil_call_function": {
			givenCtx: context.Background(),
			givenIn: func() chan string {
				ch := make(chan string)

				go func() {
					for i := 0; i < workLoad; i++ {
						ch <- strconv.Itoa(i)
					}
				}()

				return ch
			},
			expErr: merec.ErrNilCallFunc,
		},
	}

	for tcName, tc := range testCases {
		tc := tc

		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			resCh, err := merec.RunFromChan(tc.givenCtx, tc.givenIn(), tc.givenCall)
			require.ErrorIs(t, err, tc.expErr)
			require.Empty(t, resCh)
		})
	}
}

func TestRunFromChan_WithOptions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]struct {
		givenOptions []merec.CallOption[string, int]
		givenIn      chan string
		expRes       []merec.Result[int]
	}{
		"no_options": {
			givenIn: givenCh(workLoad),
			expRes:  expectedResults(),
		},
		"timeout_option": {
			givenOptions: []merec.CallOption[string, int]{
				merec.NewTimeoutOption[string, int](200 * time.Millisecond),
			},
			givenIn: givenCh(workLoad),
			expRes: []merec.Result[int]{
				merec.ErrorResult[int](fmt.Errorf("%w: %w", merec.ErrBusinessLogic, errCtxDeadline)),
				merec.ErrorResult[int](fmt.Errorf("%w: %w", merec.ErrBusinessLogic, errCtxDeadline)),
				merec.ErrorResult[int](fmt.Errorf("%w: %w", merec.ErrBusinessLogic, errCtxDeadline)),
				merec.ErrorResult[int](fmt.Errorf("%w: %w", merec.ErrBusinessLogic, errCtxDeadline)),
				merec.ErrorResult[int](fmt.Errorf("%w: %w", merec.ErrBusinessLogic, errCtxDeadline)),
			},
		},
		"timeout_option_long_enough": {
			givenOptions: []merec.CallOption[string, int]{
				merec.NewTimeoutOption[string, int](10 * time.Second),
			},
			givenIn: givenCh(workLoad),
			expRes:  expectedResults(),
		},
		"fail_fast_option_success": {
			givenOptions: []merec.CallOption[string, int]{
				merec.NewFailFastOptionOption[string, int](1),
			},
			givenIn: givenCh(workLoad),
			expRes:  expectedResults(),
		},
		"fail_fast_option_error": {
			givenOptions: []merec.CallOption[string, int]{
				merec.NewTimeoutOption[string, int](200 * time.Millisecond),
				merec.NewFailFastOptionOption[string, int](1),
			},
			givenIn: givenCh(workLoad),
			expRes: []merec.Result[int]{
				merec.ErrorResult[int](
					fmt.Errorf("%w: %w", merec.ErrBusinessLogic,
						fmt.Errorf("%w: %w", merec.ErrMustStop, errCtxDeadline),
					),
				),
			},
		},
	}

	for tcName, tc := range testCases {
		tc := tc

		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			resCh, err := merec.RunFromChan(ctx, tc.givenIn, stabCall(time.Second), tc.givenOptions...)
			require.ErrorIs(t, err, nil)

			results := make([]merec.Result[int], 0, workLoad)

			for r := range resCh {
				results = append(results, r)
			}

			require.ElementsMatch(t, results, tc.expRes)
		})
	}
}
