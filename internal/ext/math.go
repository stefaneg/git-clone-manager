package ext

// Min returns the lesser of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the greater of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
