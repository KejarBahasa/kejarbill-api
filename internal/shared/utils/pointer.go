package utils

func Pointer[T any](v T) *T {
	return &v
}

func PtrOrNil[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}

func DerefString(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
