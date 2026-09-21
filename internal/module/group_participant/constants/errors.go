package constants

import "errors"

var (
	ErrDuplicateGuestName = errors.New("duplicate guest display name")

	ErrParticipantNotFound = errors.New("participant not found")

	ErrParticipantNotClaimable = errors.New("participant is not a claimable guest")
)
