package request

import "errors"

var (
	ErrInvalidRequestBody  = errors.New("invalid request body")
	ErrInvalidPathParams   = errors.New("invalid path params")
	ErrInvalidRequestQuery = errors.New("invalid request query")
)

type ValidationError struct {
	Errors any
}

func (e *ValidationError) Error() string {
	return "validation error"
}
