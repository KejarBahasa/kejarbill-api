package entity

import "time"

type User struct {
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ID           string
	Name         string
	Username     string
	Email        string
	PasswordHash string
	Status       string
	TokenVersion int16
}
