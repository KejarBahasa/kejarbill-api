package dto

type RevealRecipientPaymentMethodParams struct {
	GroupID         string `uri:"group_id" validate:"required,uuid"`
	ParticipantID   string `uri:"participant_id" validate:"required,uuid"`
	PaymentMethodID string `uri:"payment_method_id" validate:"required,uuid"`
}
