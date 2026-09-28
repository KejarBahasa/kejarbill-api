package dto

type SettlementTimelineResponse struct {
	FromParticipant SettlementParticipantResponse `json:"from_participant"`
	ToParticipant   SettlementParticipantResponse `json:"to_participant"`
	ID              string                        `json:"id"`
	SettlementDate  string                        `json:"settlement_date"`
	Amount          int64                         `json:"amount"`
}

type SettlementParticipantResponse struct {
	ParticipantID string `json:"participant_id"`
	DisplayName   string `json:"display_name"`
}
