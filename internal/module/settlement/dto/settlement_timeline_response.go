package dto

type SettlementTimelineResponse struct {
	FromParticipant SettlementParticipantResponse `json:"from_participant"`
	ToParticipant   SettlementParticipantResponse `json:"to_participant"`
	ID              string                        `json:"id"`
	Amount          int64                         `json:"amount"`
	SettlementDate  string                        `json:"settlement_date"`
}

type SettlementParticipantResponse struct {
	ParticipantID string `json:"participant_id"`
	DisplayName   string `json:"display_name"`
}
