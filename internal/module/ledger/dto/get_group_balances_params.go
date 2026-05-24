package dto

type GetGroupBalancesParams struct {
	GroupID string `uri:"group_id" validate:"required,uuid"`
}
