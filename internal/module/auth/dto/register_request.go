package dto

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=100,alpha"`
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=100,min_special=1,min_number=1,min_alpha=1"`
}
