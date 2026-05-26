package entity

type PaginatedSettlementTimeline struct {
	Settlements []SettlementTimeline
	Page        int
	Limit       int
	TotalItems  int64
	TotalPages  int
}
