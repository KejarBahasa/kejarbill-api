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
	DiscountType       string
	DiscountValue      int64
	DiscountAmount     int64
	SubtotalAmount     int64
	TotalAmount        int64
	Version            int
}

type ExpenseDetailParticipant struct {
	ParticipantID   string
	DisplayName     string
	ParticipantType string
	ShareAmount     int64
}
