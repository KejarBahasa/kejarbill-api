package entity

import "time"

type Group struct {
	ID        string
	Name      string
	CreatedBy string
	CreatedAt time.Time
}
