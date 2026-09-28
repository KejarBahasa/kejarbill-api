package entity

import "time"

type SettlementTimeline struct {
	SettlementDate    time.Time
	ID                string
	FromParticipantID string
	FromDisplayName   string
	ToParticipantID   string
	ToDisplayName     string
	Amount            int64
}
