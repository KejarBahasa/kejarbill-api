package dto

type DeleteExpenseParams struct {
	ExpenseID string `uri:"expense_id" validate:"required,uuid"`
}
