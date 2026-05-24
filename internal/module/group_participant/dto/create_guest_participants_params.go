package dto

type CreateGuestParticipantsParams struct {
	GroupID string `uri:"group_id" validate:"required,uuid"`
}
