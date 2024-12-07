package ext

func DefaultValue[T comparable](value T, fallback T) T {
	var zero T
	if value == zero {
		return fallback
	}
	return value
}
