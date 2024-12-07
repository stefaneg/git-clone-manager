package pipe

import (
	"time"
)

func RateLimit[T any](input <-chan T, ratePerSecond int, bufferSize int) <-chan T {
	output := make(chan T, bufferSize)
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(ratePerSecond))
		defer ticker.Stop()
		for item := range input {
			<-ticker.C
			output <- item
		}
		close(output)
	}()
	return output
}
