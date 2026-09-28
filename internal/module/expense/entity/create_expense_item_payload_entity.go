package entity

type CreateExpenseItemPayload struct {
	Name          string
	ParticipantID string
	Amount        int64
}
