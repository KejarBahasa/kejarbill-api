package dto

type CheckUsernameAvailabilityRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,username"`
}
