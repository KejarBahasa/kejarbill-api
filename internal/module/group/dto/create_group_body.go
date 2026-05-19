package dto

type CreateGroupBodyRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}
