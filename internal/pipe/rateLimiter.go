package pipe

import (
	"time"
)

func RateLimit[T any](input <-chan T, output chan<- T, ratePerSecond int) {
	ticker := time.NewTicker(time.Second / time.Duration(ratePerSecond))
	defer ticker.Stop()

	for item := range input {
		<-ticker.C
		output <- item
	}
	close(output)
}
