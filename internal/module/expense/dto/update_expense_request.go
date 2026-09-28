package dto

type UpdateExpenseRequest struct {
	Title              string                           `json:"title" validate:"required,max=150"`
	Description        string                           `json:"description"`
	Currency           string                           `json:"currency" validate:"required,iso4217"`
	ExpenseDate        string                           `json:"expense_date" validate:"required"`
	PayerParticipantID string                           `json:"payer_participant_id" validate:"required,uuid"`
	SplitMethod        string                           `json:"split_method" validate:"required,oneof=equal custom itemized"`
	Version            int                              `json:"version" validate:"required,gt=0"`
	TotalAmount        int64                            `json:"total_amount"`
	ParticipantIDs     []string                         `json:"participant_ids" validate:"omitempty,min=1,dive,uuid"`
	Participants       []CreateExpenseCustomParticipant `json:"participants" validate:"omitempty,min=1,dive"`
	Items              []CreateExpenseItemizedItem      `json:"items" validate:"omitempty,min=1,dive"`
}
