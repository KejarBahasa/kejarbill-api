package request

import "errors"

var (
	ErrInvalidRequestBody   = errors.New("invalid request body")
	ErrInvalidRequestParams = errors.New("invalid request params")
	ErrInvalidRequestQuery  = errors.New("invalid request query")
)

type ValidationError struct {
	Errors any
}

func (e *ValidationError) Error() string {
	return "validation error"
}
