package constants

import "errors"

var (
	ErrParticipantsRequired = errors.New("participants required")

	ErrDuplicateParticipants = errors.New("participants contains duplicate")

	ErrInvalidExpenseDate = errors.New("invalid expense date")
)
