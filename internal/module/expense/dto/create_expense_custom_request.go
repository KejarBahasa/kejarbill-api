package dto

type CreateExpenseCustomRequest struct {
	GroupID            string                           `json:"group_id" validate:"required,uuid"`
	Title              string                           `json:"title" validate:"required,max=150"`
	Description        string                           `json:"description"`
	Currency           string                           `json:"currency" validate:"required,iso4217"`
	ExpenseDate        string                           `json:"expense_date" validate:"required"`
	PayerParticipantID string                           `json:"payer_participant_id" validate:"required,uuid"`
	Participants       []CreateExpenseCustomParticipant `json:"participants" validate:"required,min=1,dive"`
	DiscountType       string                           `json:"discount_type" validate:"omitempty,oneof=amount percentage"`
	DiscountValue      int64                            `json:"discount_value" validate:"gte=0"`
}

type CreateExpenseCustomParticipant struct {
	ParticipantID string `json:"participant_id" validate:"required,uuid"`
	ShareAmount   int64  `json:"share_amount" validate:"required,gt=0"`
}
