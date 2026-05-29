package dto

type CreateGroupBodyRequest struct {
	Name        string  `json:"name" validate:"required,max=100"`
	Description *string `json:"description" validate:"omitempty,max=255"`
}
