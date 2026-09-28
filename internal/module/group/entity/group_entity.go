package entity

import "time"

type Group struct {
	ID          string
	Name        string
	Description *string
	CreatedBy   string
	CreatedAt   time.Time
}
