package entity

import "time"

type GroupParticipant struct {
	ID              string
	GroupID         string
	UserID          *string
	DisplayName     string
	PhoneNumber     string
	ParticipantType string
	CreatedBy       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
