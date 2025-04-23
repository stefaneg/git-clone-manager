package ext

// IsChannelClosed checks if a given channel is closed. Helper function for debugging.
func IsChannelClosed[T any](ch <-chan T) bool {
	select {
	case _, ok := <-ch:
		return !ok
	default:
		return false
	}
}
