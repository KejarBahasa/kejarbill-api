package dto

type ExpenseDetailResponse struct {
	Participants []ExpenseDetailParticipantResponse `json:"participants"`
	ID           string                             `json:"id"`
	Title        string                             `json:"title"`
	Description  *string                            `json:"description"`
	Currency     string                             `json:"currency"`
	TotalAmount  int64                              `json:"total_amount"`
	ExpenseDate  string                             `json:"expense_date"`
	Payer        ExpenseDetailPayerResponse         `json:"payer"`
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
