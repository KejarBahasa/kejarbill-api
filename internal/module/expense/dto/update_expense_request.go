package dto

type UpdateExpenseRequest struct {
	Title              string                           `json:"title" validate:"required,max=150"`
	Description        string                           `json:"description"`
	Currency           string                           `json:"currency" validate:"required,iso4217"`
	ExpenseDate        string                           `json:"expense_date" validate:"required"`
	PayerParticipantID string                           `json:"payer_participant_id" validate:"required,uuid"`
	SplitMethod        string                           `json:"split_method" validate:"required,oneof=equal custom itemized"`
	Version            int                              `json:"version" validate:"required,gt=0"`
	SubtotalAmount     int64                            `json:"subtotal_amount" validate:"gte=0"`
	TotalAmount        int64                            `json:"total_amount" validate:"gte=0"`
	ParticipantIDs     []string                         `json:"participant_ids" validate:"omitempty,min=1,dive,uuid"`
	Participants       []CreateExpenseCustomParticipant `json:"participants" validate:"omitempty,min=1,dive"`
	Items              []CreateExpenseItemizedItem      `json:"items" validate:"omitempty,min=1,dive"`
	DiscountType       string                           `json:"discount_type" validate:"omitempty,oneof=amount percentage"`
	DiscountValue      int64                            `json:"discount_value" validate:"gte=0"`
}
