package errs

import "errors"

var (
	ErrBadRequest = errors.New("bad request")

	ErrUnauthorized = errors.New("unauthorized")

	ErrForbidden = errors.New("forbidden")

	ErrNotFound = errors.New("not found")

	ErrInvalidInput = errors.New("invalid input")

	ErrConflict = errors.New("conflict")

	ErrInternalServer = errors.New("internal server error")
)
