package historyin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunnerState(t *testing.T) {
	t.Run("initial state should be not stopped", func(t *testing.T) {
		var s runnerState
		assert.False(t, s.Stopped())
		assert.False(t, s.Wait(time.Duration(0)))
	})

	t.Run("Stop should cause Stopped to be true", func(t *testing.T) {
		var s runnerState
		s.Stop()
		assert.True(t, s.Stopped())
		assert.False(t, s.Wait(time.Duration(0)))
	})

	t.Run("multiple calls should be fine", func(t *testing.T) {
		var s runnerState

		for i := 0; i < 20; i++ {
			s.Stop()
			for i := 0; i < 20; i++ {
				assert.True(t, s.Stopped())
			}
			assert.False(t, s.Wait(time.Duration(0)))
		}
	})

	t.Run("Wait", func(t *testing.T) {
		t.Run("already done", func(t *testing.T) {
			var s runnerState
			s.MarkDone()
			assert.True(t, s.Wait(time.Duration(0)))
		})

		t.Run("delayed done", func(t *testing.T) {
			var s runnerState
			go func() {
				time.Sleep(10 * time.Millisecond)
				s.MarkDone()
			}()
			assert.True(t, s.Wait(time.Second))
			assert.True(t, s.Wait(0))
		})
	})

	t.Run("Drain sets IsDraining", func(t *testing.T) {
		var s runnerState
		assert.False(t, s.IsDraining())
		s.Drain()
		assert.True(t, s.IsDraining())
	})
}
