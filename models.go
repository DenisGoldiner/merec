package merec

import (
	"context"
	"fmt"
)

// CallOption is a proxy extension option that can do additional actions during the processing.
type CallOption[In, Out any] interface {
	WithOption(next Call[In, Out]) Call[In, Out]
}

// Call is the function to be executed.
type Call[In, Out any] func(context.Context, In) (Out, error)

// Result is the Call execution result. If the error happened, the value should be empty.
type Result[Out any] struct {
	value Out
	err   error
}

// String implements io.Stringer interface.
func (r Result[Out]) String() string {
	return fmt.Sprintf("value: %v, err: %v", r.value, r.err)
}

// Value returns the execution result value.
func (r Result[Out]) Value() Out {
	return r.value
}

// Err returns the execution error.
func (r Result[Out]) Err() error {
	return r.err
}

// ValueResult creates a new success result with a set value.
func ValueResult[Out any](value Out) Result[Out] {
	return Result[Out]{value: value}
}

// ErrorResult creates a new failure result with a set error.
func ErrorResult[Out any](err error) Result[Out] {
	return Result[Out]{err: err}
}
