package utils

func HasDuplicateString(values []string) bool {
	seen := make(map[string]struct{})
	for _, value := range values {
		if _, ok := seen[value]; ok {
			return true
		}

		seen[value] = struct{}{}
	}

	return false
}
