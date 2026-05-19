package constants

import "errors"

var (
	ErrParticipantsRequired = errors.New("participants required")

	ErrPayerNotIncludedInParticipants = errors.New("payer not included in participants")

	ErrPayerParticipantNotInGroup = errors.New("payer participant not found in group")

	ErrParticipantNotInGroup = errors.New("participant not found in group")

	ErrDuplicateParticipants = errors.New("participants contains duplicate")

	ErrGroupNotFound = errors.New("group not found")

	ErrForbiddenGroupAccess = errors.New("forbidden group access")

	ErrInvalidExpenseDate = errors.New("invalid expense date")
)
