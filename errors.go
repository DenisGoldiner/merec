package merec

import "errors"

// The list of supported errors.
var (
	ErrBusinessLogic = errors.New("business logic execution failed")
	ErrCtxCancel     = errors.New("root context was canceled")
	ErrCtxDeadline   = errors.New("root context's deadline passed")
	ErrNilContext    = errors.New("context must be initiated")
	ErrNilInChan     = errors.New("input channel must be initiated")
	ErrNilCallFunc   = errors.New("call function must be initiated")
	ErrMustStop      = errors.New("the processing must be interrupted")
)
