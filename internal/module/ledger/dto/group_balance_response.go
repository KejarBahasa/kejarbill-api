package dto

type GroupBalanceParticipantResponse struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type GroupBalanceResponse struct {
	FromParticipant GroupBalanceParticipantResponse `json:"from_participant"`
	ToParticipant   GroupBalanceParticipantResponse `json:"to_participant"`
	Amount          int64                           `json:"amount"`
}
