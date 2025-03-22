package channel

import (
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	input := make(chan int, 5)
	output := RateLimit(input, 200, 5)

	// Send 5 items to the input channel
	for i := 0; i < 5; i++ {
		input <- i
	}
	close(input)

	start := time.Now()
	count := 0

	// Read items from the output channel
	for range output {
		count++
	}

	duration := time.Since(start)

	// Check that we received all items
	if count != 5 {
		t.Errorf("expected 5 items, got %d", count)
	}

	// Check that the rate limit was respected (should take at least 0.025 seconds for 5 items at 200 items per second)
	if duration < 25*time.Millisecond {
		t.Errorf("rate limit not respected, took %v", duration)
	}
}
