package entity

import "time"

type ExpenseItem struct {
	Notes     *string
	CreatedAt time.Time
	UpdatedAt time.Time
	ID        string
	ExpenseID string
	Name      string
	Qty       int64
	UnitPrice int64
	Subtotal  int64
}
