package dto

type PaginationQuery struct {
	Page  int `query:"page" validate:"omitempty,min=1"`
	Limit int `query:"limit" validate:"omitempty,min=1,max=100"`
}
