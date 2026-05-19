package entity

import "time"

type Settlement struct {
	ID                string
	GroupID           string
	FromParticipantID string
	ToParticipantID   string
	Amount            int64
	Status            string
	Notes             *string
	PaidAt            time.Time
	CreatedBy         string
	CreatedAt         time.Time
	UpdatedAt         *time.Time
}
