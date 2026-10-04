package dto

type ExpenseDetailResponse struct {
	Description    *string                            `json:"description"`
	ID             string                             `json:"id"`
	Title          string                             `json:"title"`
	Currency       string                             `json:"currency"`
	ExpenseDate    string                             `json:"expense_date"`
	SubtotalAmount int64                              `json:"subtotal_amount"`
	DiscountType   string                             `json:"discount_type"`
	DiscountValue  int64                              `json:"discount_value"`
	DiscountAmount int64                              `json:"discount_amount"`
	TotalAmount    int64                              `json:"total_amount"`
	Version        int                                `json:"version"`
	Payer          ExpenseDetailPayerResponse         `json:"payer"`
	Participants   []ExpenseDetailParticipantResponse `json:"participants"`
	Items          []ExpenseDetailItemResponse        `json:"items"`
}

type ExpenseDetailPayerResponse struct {
	ParticipantID string `json:"participant_id"`
	DisplayName   string `json:"display_name"`
}

type ExpenseDetailParticipantResponse struct {
	ParticipantID   string `json:"participant_id"`
	DisplayName     string `json:"display_name"`
	ParticipantType string `json:"participant_type"`
	ShareAmount     int64  `json:"share_amount"`
}

type ExpenseDetailItemResponse struct {
	Notes        *string                                `json:"notes"`
	ID           string                                 `json:"id"`
	Name         string                                 `json:"name"`
	Qty          int64                                  `json:"qty"`
	UnitPrice    int64                                  `json:"unit_price"`
	Subtotal     int64                                  `json:"subtotal"`
	Participants []ExpenseDetailItemParticipantResponse `json:"participants"`
}

type ExpenseDetailItemParticipantResponse struct {
	ParticipantID string `json:"participant_id"`
	DisplayName   string `json:"display_name"`
	ShareAmount   int64  `json:"share_amount"`
}
