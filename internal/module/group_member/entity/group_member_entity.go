package entity

import "time"

type GroupMember struct {
	ID        string
	GroupID   string
	UserID    string
	Role      string
	Status    string
	JoinedAt  time.Time
	LeftAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
