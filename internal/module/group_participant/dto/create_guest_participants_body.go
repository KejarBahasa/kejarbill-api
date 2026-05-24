package dto

type CreateGuestParticipantItem struct {
	DisplayName string `json:"display_name" validate:"required,max=100"`
}

type CreateGuestParticipantsBody struct {
	Guests []CreateGuestParticipantItem `json:"guests" validate:"required,min=1,dive"`
}
