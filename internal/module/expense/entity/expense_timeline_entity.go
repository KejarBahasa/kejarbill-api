package entity

import "time"

type ExpenseTimeline struct {
	ID                 string
	Title              string
	Description        *string
	Currency           string
	TotalAmount        int64
	ExpenseDate        time.Time
	PayerParticipantID string
	PayerDisplayName   string
}
