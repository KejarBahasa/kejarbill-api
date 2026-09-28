package entity

import "time"

type IdempotencyKey struct {
	ResponseCode *int
	ResponseBody *string
	ExpiredAt    time.Time
	UserID       string
	Key          string
	RequestHash  string
}
