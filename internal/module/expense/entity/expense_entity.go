package entity

import "time"

type Expense struct {
	ID string

	GroupID string

	Title       string
	Description *string

	PaidByParticipantID string

	Currency string

	SubtotalAmount float64
	TaxAmount      float64
	ServiceAmount  float64
	DiscountAmount float64
	TotalAmount    float64

	SplitMethod string

	ExpenseDate time.Time

	ReceiptURL *string

	Status string

	Version int

	CreatedBy string

	CreatedAt time.Time
	UpdatedAt time.Time

	DeletedAt *time.Time
}
