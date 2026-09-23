package entity

import "time"

type MyDebtExpense struct {
	ExpenseDate     time.Time
	CreatedAt       time.Time
	ID              string
	Title           string
	ToParticipantID string
	ToDisplayName   string
	ShareAmount     int64
	NetPairAmount   int64
}
