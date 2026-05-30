package dto

type LoginRequest struct {
	Identifier string `json:"identifier" validate:"required,min=3,max=255"`
	Password   string `json:"password" validate:"required,min=8,max=128"`
}
