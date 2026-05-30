package constants

import "errors"

var (
	ErrSettlementNotFound = errors.New("settlement not found")

	ErrInvalidSettlementParticipants = errors.New("invalid settlement participants")

	ErrSettlementAmountExceeded = errors.New("settlement amount exceeded outstanding balance")
)
