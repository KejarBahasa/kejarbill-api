package entity

import "time"

type CreateExpensePayload struct {
	ExpenseDate        time.Time
	Description        *string
	GroupID            string
	Title              string
	Currency           string
	CreatedBy          string
	SplitMethod        string
	PayerParticipantID string
	Participants       []CreateExpenseParticipantPayload
	DiscountType       string
	DiscountValue      int64
	DiscountAmount     int64
	SubtotalAmount     int64
	TotalAmount        int64
}

type CreateExpenseParticipantPayload struct {
	ParticipantID string
	ShareAmount   int64
}
