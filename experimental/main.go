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

func main() {

	// Check if output is a TTY
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))

	// Number of counters
	const numCounters = 5

	// Create channels to simulate incoming numbers for each counter
	channels := make([]chan int, numCounters)
	for i := range channels {
		channels[i] = make(chan int, 1)
	}

	// Use atomic counters to keep track of the numbers
	counters := make([]int64, numCounters)

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

	go func() {
		for i := 0; i < numCounters; i++ {
			go funnelCounter(channels, i, counters)
		}
	}()

	// Goroutine to update and display all counters in place relative to the command
	if isTTY {
		fmt.Println("Entering TTY Mode")
		go func() {
			// Print initial placeholders for counters
			for i := 0; i < numCounters; i++ {
				fmt.Printf("Counter %d: 0\n", i+1)
			}

			for {
				// Move cursor up by the number of counters
				fmt.Printf("\033[%dA", numCounters)

				// Display all counters
				for i := 0; i < numCounters; i++ {
					fmt.Printf("Counter %d: %d\n", i+1, atomic.LoadInt64(&counters[i]))
				}

				time.Sleep(100 * time.Millisecond) // Refresh rate
			}
			// Block until an interrupt signal is received
		}()
		<-signalChan

		fmt.Println("\nShutting down gracefully...")
	} else {
		fmt.Println("Running for 30 seconds...then printing counters.")
		time.Sleep(time.Second * 10)
	}

}

func funnelCounter(channels []chan int, i int, counters []int64) {
	for num := range channels[i] {
		atomic.StoreInt64(&counters[i], int64(num))
	}
}
