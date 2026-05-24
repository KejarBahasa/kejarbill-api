package dto

type AddGroupMemberParams struct {
	GroupID string `uri:"group_id" validate:"required,uuid"`
}
