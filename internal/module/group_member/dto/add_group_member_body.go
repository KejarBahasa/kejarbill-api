package dto

type AddGroupMemberBody struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}
