package entity

import "time"

type SettlementDetail struct {
	PaymentMethodID   *string
	Notes             *string
	ProviderName      *string
	MethodType        *string
	PaidAt            time.Time
	ID                string
	GroupID           string
	PaymentChannel    string
	Status            string
	FromParticipantID string
	FromDisplayName   string
	ToParticipantID   string
	ToDisplayName     string
	Amount            int64
}
