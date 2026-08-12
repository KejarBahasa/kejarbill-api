package constants

import "errors"

var (
	ErrSettlementNotFound = errors.New("settlement not found")

	ErrInvalidSettlementParticipants = errors.New("invalid settlement participants")

	ErrSettlementAmountExceeded = errors.New("settlement amount exceeded outstanding balance")

	ErrInvalidSettlementPaymentMethod = errors.New("invalid payment method")

	ErrPaymentMethodMustBeEmpty = errors.New("payment method must be empty for cash payment channel")

	ErrPaymentMethodRequired = errors.New("payment method is required for bank transfer and e-wallet payment channels")

	ErrPaymentMethodNotOwned = errors.New("payment method is not owned by the recipient")

	ErrInvalidPaymentMethodType = errors.New("invalid payment method type for the selected payment channel")

	ErrIdempotencyKeyRequired = errors.New("idempotency key is required")

	ErrIdempotencyKeyConflict = errors.New("idempotency key has already been used with a different request")

	ErrIdempotencyResponseUnavailable = errors.New("idempotency response is unavailable")
)
