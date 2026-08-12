package constants

import "errors"

var (
	ErrPaymentMethodNotFound = errors.New("payment method not found")

	ErrPaymentMethodInactive = errors.New("payment method is inactive")
)
