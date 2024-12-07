package counter

import "sync"

type Counter struct {
	addChan   chan int
	countChan chan int
	wg        sync.WaitGroup
}

// NewCounter creates and initializes a new Counter
func NewCounter() *Counter {
	c := &Counter{
		addChan:   make(chan int),
		countChan: make(chan int),
		wg:        sync.WaitGroup{},
	}

	go c.receiveCounts()
	return c
}

func (c *Counter) receiveCounts() {
	var total int
	for {
		select {
		case add := <-c.addChan:
			total += add
			c.wg.Done()
		case c.countChan <- total:
			// Sends the current total when requested
		}
	}
}

// Add adds a value to the counter safely
func (c *Counter) Add(value int) {
	c.wg.Add(1)
	c.addChan <- value
}

// Count returns the current count safely
func (c *Counter) Count() int {
	c.wg.Wait()
	return <-c.countChan
}
