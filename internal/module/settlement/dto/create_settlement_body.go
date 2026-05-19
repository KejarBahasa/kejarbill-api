package dto

type CreateSettlementBody struct {
	FromParticipantID string `json:"from_participant_id" validate:"required,uuid"`
	ToParticipantID   string `json:"to_participant_id" validate:"required,uuid"`
	Amount            int64  `json:"amount" validate:"required,gt=0"`
	Notes             string `json:"notes"`
	PaidAt            string `json:"paid_at" validate:"required"`
}
