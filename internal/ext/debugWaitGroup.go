/*
Package ext is "language extensions", functionality that in a perfect world would be part of the golang standard library
*/
package ext

import (
	"fmt"
	logger "gcm/internal/log"
	"sync"
	"sync/atomic"
)

type DebugWaitGroup struct {
	wg      sync.WaitGroup
	counter int64
}

func (d *DebugWaitGroup) Add(delta int) {
	atomic.AddInt64(&d.counter, int64(delta))
	fmt.Printf("Add(%d): counter = %d\n", delta, atomic.LoadInt64(&d.counter))
	d.wg.Add(delta)
}

func (d *DebugWaitGroup) Done() {
	atomic.AddInt64(&d.counter, -1)
	fmt.Printf("Done(): counter = %d\n", atomic.LoadInt64(&d.counter))
	d.wg.Done()
}

func (d *DebugWaitGroup) Wait() {
	d.wg.Wait()
	if atomic.LoadInt64(&d.counter) != 0 {
		fmt.Printf("Warning: counter = %d at Wait()\n", atomic.LoadInt64(&d.counter))
	}
}

func (d *DebugWaitGroup) AssertStillWaiting() {
	if atomic.LoadInt64(&d.counter) == 0 {
		logger.Log.Errorf("COUNTER DOWN TO = %d ! \n", atomic.LoadInt64(&d.counter))
	}
}
