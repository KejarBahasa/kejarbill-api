package security

import "time"

type Payload struct {
	TokenID   string
	UserID    string
	ExpiredAt time.Time
}
