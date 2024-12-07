package counter

import (
	"testing"
)

func TestCounter(t *testing.T) {
	t.Run("InitialCountIsZero", func(t *testing.T) {
		counter := NewCounter()
		if got := counter.Count(); got != 0 {
			t.Errorf("Expected initial count to be 0, got %d", got)
		}
	})

	t.Run("SingleAdd", func(t *testing.T) {
		counter := NewCounter()
		counter.Add(1)
		if got := counter.Count(); got != 1 {
			t.Errorf("Expected count to be 1 after adding 1, got %d", got)
		}
	})

	t.Run("MultipleAdds", func(t *testing.T) {
		counter := NewCounter()
		for i := 0; i < 10; i++ {
			counter.Add(1)
		}
		if got := counter.Count(); got != 10 {
			t.Errorf("Expected count to be 10 after adding 1 ten times, got %d", got)
		}
	})

	t.Run("ConcurrentAdds", func(t *testing.T) {
		counter := NewCounter()
		const goroutines = 10
		const addsPerGoroutine = 10

		done := make(chan struct{}, goroutines)
		for i := 0; i < goroutines; i++ {
			go func() {
				for j := 0; j < addsPerGoroutine; j++ {
					counter.Add(1)
				}
				done <- struct{}{}
			}()
		}

		// Wait for all goroutines to finish
		for i := 0; i < goroutines; i++ {
			<-done
		}

		expected := goroutines * addsPerGoroutine
		if got := counter.Count(); got != expected {
			t.Errorf("Expected count to be %d after concurrent adds, got %d", expected, got)
		}

		counter.Add(1)
		if got := counter.Count(); got != expected+1 {
			t.Errorf("Expected count to be %d after concurrent adds, got %d", expected+1, got)
		}

	})
}
