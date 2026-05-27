package dto

type CreateExpenseItemizedRequest struct {
	GroupID            string                      `json:"group_id" validate:"required,uuid"`
	Title              string                      `json:"title" validate:"required,max=150"`
	Description        string                      `json:"description"`
	Currency           string                      `json:"currency" validate:"required,iso4217"`
	ExpenseDate        string                      `json:"expense_date" validate:"required"`
	PayerParticipantID string                      `json:"payer_participant_id" validate:"required,uuid"`
	Items              []CreateExpenseItemizedItem `json:"items" validate:"required,min=1,dive"`
}

type CreateExpenseItemizedItem struct {
	Name          string `json:"name" validate:"required,max=150"`
	ParticipantID string `json:"participant_id" validate:"required,uuid"`
	Notes         string `json:"notes"`
	Qty           int64  `json:"qty" validate:"required,gt=0"`
	UnitPrice     int64  `json:"unit_price" validate:"required,gt=0"`
}
