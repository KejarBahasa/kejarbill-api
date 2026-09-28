package dto

type SettlementDetailResponse struct {
	Notes           *string                            `json:"notes"`
	PaymentMethod   *SettlementDetailPaymentMethodInfo `json:"payment_method,omitempty"`
	ID              string                             `json:"id"`
	GroupID         string                             `json:"group_id"`
	PaymentChannel  string                             `json:"payment_channel"`
	PaidAt          string                             `json:"paid_at"`
	Status          string                             `json:"status"`
	FromParticipant SettlementDetailParticipantInfo    `json:"from_participant"`
	ToParticipant   SettlementDetailParticipantInfo    `json:"to_participant"`
	Amount          int64                              `json:"amount"`
}

type SettlementDetailParticipantInfo struct {
	ParticipantID string `json:"participant_id"`
	DisplayName   string `json:"display_name"`
}

type SettlementDetailPaymentMethodInfo struct {
	ID           string `json:"id"`
	MethodType   string `json:"method_type"`
	ProviderName string `json:"provider_name"`
}
