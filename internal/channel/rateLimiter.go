package channel

import (
	"gcm/internal/ext"
	"time"
)

func RateLimit[T any](input <-chan T, ratePerSecond int, bufferSize int) <-chan T {
	output := make(chan T, bufferSize)
	go func() {
		ticker := time.NewTicker(time.Duration(ext.MaxI64(int64(time.Second/time.Duration(ratePerSecond)), 1)))
		defer ticker.Stop()
		for item := range input {
			<-ticker.C
			output <- item
		}
		close(output)
	}()
	return output
}
