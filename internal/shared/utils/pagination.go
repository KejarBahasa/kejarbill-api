package utils

func NormalizePagination(page int, limit int) (int, int) {
	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	return page, limit
}

func CalculateOffset(page int, limit int) int {
	return (page - 1) * limit
}

func CalculateTotalPages(total int64, limit int) int {
	if total == 0 {
		return 0
	}

	return int((total + int64(limit) - 1) / int64(limit))
}
