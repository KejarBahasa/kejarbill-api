package dto

type GetExpenseDetailParams struct {
	ExpenseID string `uri:"expense_id" validate:"required,uuid"`
}
