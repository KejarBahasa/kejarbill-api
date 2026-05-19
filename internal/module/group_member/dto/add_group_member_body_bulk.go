package dto

type AddGroupMembersBulkBody struct {
	UserIDs []string `json:"user_ids" validate:"required,min=1,dive,uuid"`
}
