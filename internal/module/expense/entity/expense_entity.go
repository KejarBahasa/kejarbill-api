package entity

import "time"

type Expense struct {
	DeletedAt           *time.Time
	Description         *string
	ReceiptURL          *string
	ExpenseDate         time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
	ID                  string
	GroupID             string
	Title               string
	PaidByParticipantID string
	Currency            string
	SplitMethod         string
	Status              string
	CreatedBy           string
	SubtotalAmount      int64
	TaxAmount           int64
	ServiceAmount       int64
	DiscountAmount      int64
	TotalAmount         int64
	Version             int
}
