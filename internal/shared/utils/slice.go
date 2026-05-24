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

func UniqueStrings(values []string) []string {
	keys := make(map[string]struct{})
	result := make([]string, 0)

	for _, value := range values {
		if _, exists := keys[value]; exists {
			continue
		}

		keys[value] = struct{}{}

		result = append(result, value)
	}

	return result
}
