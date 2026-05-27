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
	TotalAmount        int64
}

type CreateExpenseParticipantPayload struct {
	ParticipantID string
	ShareAmount   int64
}
