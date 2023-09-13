package historyin

import (
	"errors"
	"time"
)

var (
	errRunnerStopped = errors.New("runner has been stopped")
)

type runnerContext struct {
	runnerState *runnerState
}

func (ctx runnerContext) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (ctx runnerContext) Done() <-chan struct{} {
	return ctx.runnerState.stopChan
}

func (ctx runnerContext) Err() error {
	if ctx.runnerState.Stopped() {
		return errRunnerStopped
	}
	return nil
}

func (ctx runnerContext) Value(key interface{}) interface{} {
	return nil
}
