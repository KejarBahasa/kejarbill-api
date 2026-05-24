package constants

import "errors"

var (
	ErrAlreadyMember = errors.New("user already member")

	ErrUserNotFound = errors.New("one or more users not found")

	ErrSomeUsersAlreadyMember = errors.New("one or more users already member")
)
