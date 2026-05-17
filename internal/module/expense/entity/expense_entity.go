package entity

import "time"

type Expense struct {
	ID string

	GroupID string

	Title       string
	Description *string

	PaidByParticipantID string

	Currency string

	SubtotalAmount int64
	TaxAmount      int64
	ServiceAmount  int64
	DiscountAmount int64
	TotalAmount    int64

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
