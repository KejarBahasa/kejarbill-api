package entity

import "time"

type ExpenseItem struct {
	Participants []ExpenseItemParticipant
	Notes        *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ID           string
	ExpenseID    string
	Name         string
	Qty          int64
	UnitPrice    int64
	Subtotal     int64
}

type ExpenseItemParticipant struct {
	ID            string
	ExpenseItemID string
	ParticipantID string
	DisplayName   string
	ShareAmount   int64
}
