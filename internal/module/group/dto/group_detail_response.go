package dto

type GroupDetailResponse struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Description       *string `json:"description"`
	CreatedAt         string  `json:"created_at"`
	TotalMembers      int64   `json:"total_members"`
	TotalParticipants int64   `json:"total_participants"`
	TotalExpenses     int64   `json:"total_expenses"`
}
