package security

import "time"

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type Payload struct {
	ExpiredAt    time.Time
	TokenID      string
	UserID       string
	TokenType    string
	TokenVersion int16
}
