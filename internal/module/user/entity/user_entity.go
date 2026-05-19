package entity

import "time"

type User struct {
	ID        string
	Name      string
	Username  string
	Email     string
	AvatarUrl *string
	Status    string
	CreatedAt time.Time
	UpdatedAt *time.Time
}
