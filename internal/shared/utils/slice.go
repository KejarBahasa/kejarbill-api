package utils

func HasDuplicateString(values []string) bool {
	seen := make(map[string]bool)
	for _, value := range values {
		if seen[value] {
			return true
		}

		seen[value] = true
	}

	return false
}
