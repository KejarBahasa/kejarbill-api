package dto

type ActivityTimelineResponse struct {
	Expense    *ActivityExpenseResponse    `json:"expense,omitempty"`
	Settlement *ActivitySettlementResponse `json:"settlement,omitempty"`
	Type       string                      `json:"type"`
	CreatedAt  string                      `json:"created_at"`
}

type ActivityExpenseResponse struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	PayerDisplayName string `json:"payer_display_name"`
	TotalAmount      int64  `json:"total_amount"`
}

type ActivitySettlementResponse struct {
	ID              string `json:"id"`
	FromDisplayName string `json:"from_display_name"`
	ToDisplayName   string `json:"to_display_name"`
	Amount          int64  `json:"amount"`
}
