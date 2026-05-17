package entity

import "time"

type AccountLedger struct {
	ID string

	GroupID string

	FromParticipantID string
	ToParticipantID   string

	SourceType string
	SourceID   string

	Amount int64

	CreatedAt time.Time
}
