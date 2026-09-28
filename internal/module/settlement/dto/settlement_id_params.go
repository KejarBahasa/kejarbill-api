package dto

type SettlementIDParams struct {
	SettlementID string `uri:"settlement_id" validate:"required,uuid"`
}
