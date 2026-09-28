package dto

type MyDebtsResponse struct {
	Debts           []MyDebtResponse `json:"debts"`
	TotalAmount     int64            `json:"total_amount"`
	PaidAmount      int64            `json:"paid_amount"`
	RemainingAmount int64            `json:"remaining_amount"`
}

type MyDebtResponse struct {
	ToParticipant   DebtParticipantResponse `json:"to_participant"`
	TotalAmount     int64                   `json:"total_amount"`
	PaidAmount      int64                   `json:"paid_amount"`
	RemainingAmount int64                   `json:"remaining_amount"`
	Status          string                  `json:"status"`
	Expenses        []MyDebtExpenseResponse `json:"expenses"`
}

type DebtParticipantResponse struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type MyDebtExpenseResponse struct {
	ExpenseID       string `json:"expense_id"`
	Title           string `json:"title"`
	ExpenseDate     string `json:"expense_date"`
	Amount          int64  `json:"amount"`
	PaidAmount      int64  `json:"paid_amount"`
	RemainingAmount int64  `json:"remaining_amount"`
	Status          string `json:"status"`
}
