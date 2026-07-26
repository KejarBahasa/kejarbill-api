package dto

type CreateSettlementBody struct {
	FromParticipantID string `json:"from_participant_id" validate:"required,uuid"`
	ToParticipantID   string `json:"to_participant_id" validate:"required,uuid"`
	PaymentChannel    string `json:"payment_channel" validate:"required,oneof=cash bank_transfer ewallet"`
	PaymentMethodID   string `json:"payment_method_id" validate:"omitempty,uuid"`
	Notes             string `json:"notes"`
	PaidAt            string `json:"paid_at" validate:"required"`
	Amount            int64  `json:"amount" validate:"required,gt=0"`
}
