package dto

type ParticipantResponse struct {
	ID              string  `json:"id"`
	UserID          *string `json:"user_id"`
	ParticipantType string  `json:"participant_type"`
	DisplayName     string  `json:"display_name"`
	Role            *string `json:"role"`
	IsSelf          bool    `json:"is_self"`
}
