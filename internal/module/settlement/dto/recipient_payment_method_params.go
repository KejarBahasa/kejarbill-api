package dto

type RecipientPaymentMethodParams struct {
	GroupID       string `uri:"group_id" validate:"required,uuid"`
	ParticipantID string `uri:"participant_id" validate:"required,uuid"`
}
