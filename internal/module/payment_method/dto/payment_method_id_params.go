package dto

type PaymentMethodIDParams struct {
	PaymentMethodID string `uri:"payment_method_id" validate:"required,uuid"`
}
