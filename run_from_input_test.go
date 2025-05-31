package merec_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/DenisGoldiner/merec"
)

func TestRunFromInput(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		givenCtx  func(context.Context) (context.Context, context.CancelFunc)
		givenIn   string
		givenCall merec.Call[string, int]
		expRes    merec.Result[int]
	}{
		"success": {
			givenIn:   "1",
			givenCall: stabCall(time.Second),
			expRes:    merec.ValueResult(1),
		},
		"timeout": {
			givenCtx: func(ctx context.Context) (context.Context, context.CancelFunc) {
				//nolint:govet // reason: that is a test, we don't care.
				ctx, _ = context.WithTimeout(ctx, 100*time.Microsecond)
				return ctx, nil
			},
			givenIn:   "1",
			givenCall: stabCall(time.Second),
			expRes:    merec.ErrorResult[int](errCtxDeadline),
		},
		"cancel": {
			givenCtx: func(ctx context.Context) (context.Context, context.CancelFunc) {
				return context.WithTimeout(ctx, 10*time.Second)
			},
			givenIn:   "1",
			givenCall: stabCall(time.Second),
			expRes:    merec.ErrorResult[int](errCtxCancel),
		},
		"business_error": {
			givenIn:   "qwerty",
			givenCall: stabCall(time.Second),
			expRes:    merec.ErrorResult[int](merec.ErrBusinessLogic),
		},
	}

	for tcName, tc := range testCases {
		tc := tc

		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			tcCtx, tcCtxCsl := ctx, context.CancelFunc(nil)
			if tc.givenCtx != nil {
				tcCtx, tcCtxCsl = tc.givenCtx(ctx)
			}

			if tcCtxCsl != nil {
				tcCtxCsl()
			}

			resCh, err := merec.RunFromInput(tcCtx, tc.givenIn, tc.givenCall)
			require.NoError(t, err)

			res := <-resCh
			require.ErrorIs(t, res.Err(), tc.expRes.Err())
			require.Equal(t, tc.expRes.Value(), res.Value())
		})
	}
}

func TestRunFromInput_ValidationFail(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		givenCtx  context.Context
		givenIn   string
		givenCall merec.Call[string, int]
		expErr    error
	}{
		"nil_ctx": {
			givenIn:   "1",
			givenCall: stabCall(time.Second),
			expErr:    merec.ErrNilContext,
		},
		"nil_call_function": {
			givenCtx: context.Background(),
			givenIn:  "1",
			expErr:   merec.ErrNilCallFunc,
		},
	}

	for tcName, tc := range testCases {
		tc := tc

		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			resCh, err := merec.RunFromInput(tc.givenCtx, tc.givenIn, tc.givenCall)
			require.ErrorIs(t, err, tc.expErr)
			require.Nil(t, resCh)
		})
	}
}

func TestRunFromInput_WithOptions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]struct {
		givenOptions []merec.CallOption[string, int]
		givenIn      string
		expRes       merec.Result[int]
	}{
		"no_options": {
			givenIn: "1",
			expRes:  merec.ValueResult(1),
		},
		"timeout_option": {
			givenOptions: []merec.CallOption[string, int]{
				merec.NewTimeoutOption[string, int](200 * time.Millisecond),
			},
			givenIn: "1",
			expRes:  merec.ErrorResult[int](errCtxDeadline),
		},
		"timeout_option_long_enough": {
			givenOptions: []merec.CallOption[string, int]{
				merec.NewTimeoutOption[string, int](10 * time.Second),
			},
			givenIn: "1",
			expRes:  merec.ValueResult(1),
		},
		"fail_fast_option_success": {
			givenOptions: []merec.CallOption[string, int]{
				merec.NewFailFastOptionOption[string, int](1),
			},
			givenIn: "1",
			expRes:  merec.ValueResult(1),
		},
		"fail_fast_option_error": {
			givenOptions: []merec.CallOption[string, int]{
				merec.NewFailFastOptionOption[string, int](1),
			},
			givenIn: "qwerty",
			expRes:  merec.ErrorResult[int](merec.ErrMustStop),
		},
	}

	for tcName, tc := range testCases {
		tc := tc

		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			resCh, err := merec.RunFromInput(ctx, tc.givenIn, stabCall(time.Second), tc.givenOptions...)
			require.ErrorIs(t, err, nil)

			res := <-resCh
			require.ErrorIs(t, res.Err(), tc.expRes.Err())
			require.Equal(t, tc.expRes.Value(), res.Value())
		})
	}
}

func TestRunFromInputWithOptions_ValidationFail(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		givenCtx  context.Context
		givenIn   string
		givenCall merec.Call[string, int]
		expErr    error
	}{
		"nil_ctx": {
			givenIn:   "1",
			givenCall: stabCall(time.Second),
			expErr:    merec.ErrNilContext,
		},
		"nil_call_function": {
			givenCtx: context.Background(),
			givenIn:  "1",
			expErr:   merec.ErrNilCallFunc,
		},
	}

	for tcName, tc := range testCases {
		tc := tc

		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			resCh, err := merec.RunFromInput(tc.givenCtx, tc.givenIn, tc.givenCall)
			require.ErrorIs(t, err, tc.expErr)
			require.Nil(t, resCh)
		})
	}
}
