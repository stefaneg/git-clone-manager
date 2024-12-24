package main

import (
	"fmt"
	"golang.org/x/term"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

// CounterStore handles storing and retrieving counter values
type CounterStore struct {
	counters []int64
}

func NewCounterStore(numCounters int) *CounterStore {
	return &CounterStore{
		counters: make([]int64, numCounters),
	}
}

func (cs *CounterStore) UpdateCounter(index int, value int64) {
	atomic.StoreInt64(&cs.counters[index], value)
}

func (cs *CounterStore) GetCounters() []int64 {
	n := len(cs.counters)
	values := make([]int64, n)
	for i := 0; i < n; i++ {
		values[i] = atomic.LoadInt64(&cs.counters[i])
	}
	return values
}

// Renderer handles rendering counters in different modes
type Renderer struct {
	store       *CounterStore
	isTTY       bool
	numCounters int
}

func NewRenderer(store *CounterStore, isTTY bool, numCounters int) *Renderer {
	return &Renderer{
		store:       store,
		isTTY:       isTTY,
		numCounters: numCounters,
	}
}

func (r *Renderer) Render() {
	if r.isTTY {
		r.renderTTY()
	} else {
		r.renderNonTTY()
	}
}

func (r *Renderer) renderTTY() {
	// Initial placeholder rendering to create space for counters
	r.printCounters()

	for {
		// Move cursor up by the number of counters
		fmt.Printf("\033[%dA", r.numCounters)
		r.printCounters()
		time.Sleep(100 * time.Millisecond) // Refresh rate
	}
}

func (r *Renderer) printCounters() {
	counters := r.store.GetCounters()
	for i := 0; i < r.numCounters; i++ {
		fmt.Printf("Counter %d: %d\n", i+1, counters[i])
	}
}

func (r *Renderer) renderNonTTY() {
	for {
		fmt.Println("---")
		r.printCounters()
		time.Sleep(1 * time.Second) // Refresh rate
	}
}

func main() {
	// Number of counters
	const numCounters = 5

	// Check if output is a TTY
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))

	// Create CounterStore
	store := NewCounterStore(numCounters)

	// Create channels to simulate incoming numbers for each counter
	channels := make([]chan int, numCounters)
	for i := range channels {
		channels[i] = make(chan int, 1)
	}

	// Signal handling to gracefully shut down
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Goroutines to simulate incoming numbers for each counter
	for i := 0; i < numCounters; i++ {
		go func(idx int) {
			for j := 1; ; j++ {
				channels[idx] <- j
				time.Sleep(1 * time.Second) // Simulate a delay for incoming numbers
			}
		}(i)
	}

	// Goroutines to update the counter store, one per channel
	for i := 0; i < numCounters; i++ {
		go func(idx int) {
			for num := range channels[idx] {
				store.UpdateCounter(idx, int64(num))
			}
		}(i)
	}

	// Create Renderer and start rendering
	renderer := NewRenderer(store, isTTY, numCounters)
	go renderer.Render()

	// Block until an interrupt signal is received
	<-signalChan

	fmt.Println("\nShutting down gracefully...")
}
