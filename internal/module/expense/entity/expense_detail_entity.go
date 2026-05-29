package entity

import "time"

type ExpenseDetail struct {
	Participants       []ExpenseDetailParticipant
	Items              []ExpenseItem
	ID                 string
	GroupID            string
	Title              string
	Description        *string
	Currency           string
	ExpenseDate        time.Time
	PayerParticipantID string
	PayerDisplayName   string
	TotalAmount        int64
}

type ExpenseDetailParticipant struct {
	ParticipantID   string
	DisplayName     string
	ParticipantType string
	ShareAmount     int64
}
