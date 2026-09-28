package dto

type GroupSummaryResponse struct {
	GroupTotal    int64 `json:"group_total"`
	MyTotalPaid   int64 `json:"my_total_paid"`
	MyTotalDebt   int64 `json:"my_total_debt"`
	MyTotalCredit int64 `json:"my_total_credit"`
}
