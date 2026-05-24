package dto

type ExpenseTimelineResponse struct {
	Payer       ExpenseTimelinePayerResponse `json:"payer"`
	ID          string                       `json:"id"`
	Title       string                       `json:"title"`
	Description *string                      `json:"description"`
	Currency    string                       `json:"currency"`
	TotalAmount int64                        `json:"total_amount"`
	ExpenseDate string                       `json:"expense_date"`
}

type ExpenseTimelinePayerResponse struct {
	ParticipantID string `json:"participant_id"`
	DisplayName   string `json:"display_name"`
}
