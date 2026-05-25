package entity

import "time"

type SettlementTimeline struct {
	ID                string
	Amount            int64
	Currency          string
	SettlementDate    time.Time
	FromParticipantID string
	FromDisplayName   string
	ToParticipantID   string
	ToDisplayName     string
}
