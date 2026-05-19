package constants

import "errors"

var (
	ErrInvalidSettlementParticipants = errors.New("invalid settlement participants")

	ErrSettlementAmountExceeded = errors.New("settlement amount exceeded outstanding balance")
)
