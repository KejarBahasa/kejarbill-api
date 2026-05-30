package entity

import "time"

type Settlement struct {
	PaymentMethodID   *string
	Notes             *string
	UpdatedAt         *time.Time
	PaidAt            time.Time
	CreatedAt         time.Time
	ID                string
	GroupID           string
	FromParticipantID string
	ToParticipantID   string
	PaymentChannel    string
	Status            string
	CreatedBy         string
	Amount            int64
}
