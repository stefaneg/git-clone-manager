package ext

// Min returns the lesser of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MaxI returns the greater of two integers
func MaxI(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MaxI64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
