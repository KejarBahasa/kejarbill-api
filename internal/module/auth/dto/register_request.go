package dto

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=1,max=100,alphanumspace"`
	Username string `json:"username" validate:"required,min=3,max=50,username"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=128,password_strength"`
}
