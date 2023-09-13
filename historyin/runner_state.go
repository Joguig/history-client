package historyin

import (
	"context"
	"sync"
	"time"
)

// runnerState is a struct to aid in stopping go routines that can only be stopped
// once.
type runnerState struct {
	initSync sync.Once

	done     bool
	doneLock sync.Mutex
	doneChan chan struct{}

	// notifies runner to not listen on channels triggering threshold breaches
	draining bool

	stopped  bool
	stopLock sync.Mutex
	stopChan chan struct{}
}

func (rs *runnerState) init() {
	rs.initSync.Do(func() {
		rs.doneChan = make(chan struct{}, 1)
		rs.stopChan = make(chan struct{}, 1)
	})
}

func (rs *runnerState) Drain() {
	rs.draining = true
}

func (rs *runnerState) IsDraining() bool {
	return rs.draining
}

// Stop stops the runner
func (rs *runnerState) Stop() {
	rs.init()

	rs.stopLock.Lock()
	if !rs.stopped {
		rs.stopChan <- struct{}{}
		rs.stopped = true
	}
	rs.stopLock.Unlock()
}

// Stopped returns true if stopped
func (rs *runnerState) Stopped() bool {
	rs.init()

	return rs.stopped
}

// Called by the runner to mark state as done
func (rs *runnerState) MarkDone() {
	rs.init()

	rs.doneLock.Lock()
	if !rs.done {
		rs.doneChan <- struct{}{}
		rs.done = true
	}
	rs.doneLock.Unlock()
}

// Wait waits until runner acks stop
func (rs *runnerState) Wait(timeout time.Duration) bool {
	rs.init()

	if rs.done {
		return true
	}

	select {
	case <-time.NewTimer(timeout).C:
		return false
	case <-rs.doneChan:
		close(rs.doneChan)
		return true
	}
}

func (rs *runnerState) Context() context.Context {
	rs.init()

	return runnerContext{rs}
}

func (rs *runnerState) Sleep(timeout time.Duration) {
	select {
	case <-time.NewTimer(timeout).C:
	case <-rs.stopChan:
	}
}
