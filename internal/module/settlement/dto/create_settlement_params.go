package dto

type CreateSettlementParams struct {
	GroupID string `uri:"group_id" validate:"required,uuid"`
}
