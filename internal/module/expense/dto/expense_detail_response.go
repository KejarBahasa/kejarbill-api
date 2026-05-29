package dto

type ExpenseDetailResponse struct {
	Description  *string                            `json:"description"`
	ID           string                             `json:"id"`
	Title        string                             `json:"title"`
	Currency     string                             `json:"currency"`
	ExpenseDate  string                             `json:"expense_date"`
	TotalAmount  int64                              `json:"total_amount"`
	Payer        ExpenseDetailPayerResponse         `json:"payer"`
	Participants []ExpenseDetailParticipantResponse `json:"participants"`
	Items        []ExpenseItemResponse              `json:"items"`
}

type ExpenseDetailPayerResponse struct {
	ParticipantID string `json:"participant_id"`
	DisplayName   string `json:"display_name"`
}

type ExpenseDetailParticipantResponse struct {
	ParticipantID string `json:"participant_id"`
	DisplayName   string `json:"display_name"`
	ShareAmount   int64  `json:"share_amount"`
}

type ExpenseItemResponse struct {
	Notes     *string `json:"notes"`
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Qty       int64   `json:"qty"`
	UnitPrice int64   `json:"unit_price"`
	Subtotal  int64   `json:"subtotal"`
}
