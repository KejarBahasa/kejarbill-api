package dto

type ClaimParticipantBody struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}
