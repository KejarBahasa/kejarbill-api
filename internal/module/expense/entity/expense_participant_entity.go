package entity

import "time"

type ExpenseParticipant struct {
	ID            string
	ExpenseID     string
	ParticipantID string
	ShareAmount   float64
	CreatedAt     time.Time
}
