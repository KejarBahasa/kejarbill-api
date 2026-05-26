package entity

import "time"

type GroupDetail struct {
	Description       *string
	CreatedAt         time.Time
	ID                string
	Name              string
	TotalMembers      int64
	TotalParticipants int64
	TotalExpenses     int64
}
