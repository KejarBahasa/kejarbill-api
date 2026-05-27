package entity

import "time"

type ActivityTimeline struct {
	Expense    *ActivityExpense
	Settlement *ActivitySettlement
	CreatedAt  time.Time
	Type       string
}

type ActivityExpense struct {
	ID               string
	Title            string
	PayerDisplayName string
	TotalAmount      int64
}

type ActivitySettlement struct {
	ID              string
	FromDisplayName string
	ToDisplayName   string
	Amount          int64
}
