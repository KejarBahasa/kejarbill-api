package entity

type PaginatedExpenseTimeline struct {
	Expenses   []ExpenseTimeline
	Page       int
	Limit      int
	TotalItems int64
	TotalPages int
}
