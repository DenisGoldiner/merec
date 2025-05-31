package merec_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/DenisGoldiner/merec"
)

func TestCheckContext(t *testing.T) {
	t.Parallel()

	t.Run("active", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		actErr := merec.CheckContext(ctx)
		require.NoError(t, actErr)
	})
	t.Run("deadline", func(t *testing.T) {
		t.Parallel()
		ctx, ctxCsl := context.WithDeadline(context.Background(), time.Now())
		defer ctxCsl()
		actErr := merec.CheckContext(ctx)
		require.ErrorIs(t, actErr, merec.ErrCtxDeadline)
	})
	t.Run("timeout", func(t *testing.T) {
		t.Parallel()
		ctx, ctxCsl := context.WithTimeout(context.Background(), time.Nanosecond)
		defer ctxCsl()
		actErr := merec.CheckContext(ctx)
		require.ErrorIs(t, actErr, merec.ErrCtxDeadline)
	})
	t.Run("cancel", func(t *testing.T) {
		t.Parallel()
		ctx, ctxCsl := context.WithCancel(context.Background())
		ctxCsl()
		actErr := merec.CheckContext(ctx)
		require.ErrorIs(t, actErr, merec.ErrCtxCancel)
	})
}
