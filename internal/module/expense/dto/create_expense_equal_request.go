package dto

type CreateExpenseEqualRequest struct {
	ParticipantIDs     []string `json:"participant_ids" validate:"required,min=1,dive,uuid"`
	GroupID            string   `json:"group_id" validate:"required,uuid"`
	Title              string   `json:"title" validate:"required,max=150"`
	Description        string   `json:"description"`
	PayerParticipantID string   `json:"payer_participant_id" validate:"required,uuid"`
	Currency           string   `json:"currency" validate:"required,alpha,len=3"`
	ExpenseDate        string   `json:"expense_date" validate:"required"`
	TotalAmount        int64    `json:"total_amount" validate:"gt=0"`
}
