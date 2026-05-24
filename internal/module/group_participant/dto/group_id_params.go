package dto

type GroupIDParams struct {
	GroupID string `uri:"group_id" validate:"required,uuid"`
}
