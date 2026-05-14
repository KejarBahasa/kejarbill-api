package dto

type CreateExpenseEqualRequest struct {
	ParticipantIDs     []string `json:"participant_ids"`
	GroupID            string   `json:"group_id" validate:"required,uuid"`
	Title              string   `json:"title" validate:"required,max=150"`
	Description        string   `json:"description"`
	PayerParticipantID string   `json:"payer_participant_id"`
	Currency           string   `json:"currency"`
	ExpenseDate        string   `json:"expense_date"`
	TotalAmount        float64  `json:"total_amount"`
}
