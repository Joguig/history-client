package historyin

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatch(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		assertNoBreach := func(t *testing.T, b *batch) {
			var thresholdBreached bool
			select {
			case <-b.ThresholdBreach():
				thresholdBreached = true
			default:
			}

			assert.False(t, thresholdBreached, "failure should not breach threshold")
		}

		t.Run("success and breach is crossed", func(t *testing.T) {
			b := batch{}
			assert.NoError(t, b.Add(&Audit{}))
			<-b.ThresholdBreach()
			assertNoBreach(t, &b)
		})

		t.Run("successive calls should not send breach", func(t *testing.T) {
			b := batch{}
			assert.NoError(t, b.Add(&Audit{}))
			<-b.ThresholdBreach()
			assertNoBreach(t, &b)
			assert.NoError(t, b.Add(&Audit{}))
			assertNoBreach(t, &b)
		})

		t.Run("MarkThresholdBreachRead should reset breanch", func(t *testing.T) {
			b := batch{}
			assert.NoError(t, b.Add(&Audit{}))
			<-b.ThresholdBreach()
			assertNoBreach(t, &b)

			b.MarkThresholdBreachRead()
			assertNoBreach(t, &b)

			assert.NoError(t, b.Add(&Audit{}))
			<-b.ThresholdBreach()
			assertNoBreach(t, &b)
		})

		t.Run("invalid audit", func(t *testing.T) {
			b := batch{}
			err := b.Add(&Audit{UUID: "abc"})
			require.Error(t, err)
			assert.True(t, strings.Contains(err.Error(), "Invalid UUID"))
			assertNoBreach(t, &b)
		})

		t.Run("uses UUID as record key", func(t *testing.T) {
			b := batch{}
			err := b.Add(&Audit{})
			require.NoError(t, err)
			assert.Equal(t, 1, len(b.records))
			for _, record := range b.records {
				assert.NotEmpty(t, record.Key)
				assert.Equal(t, 36, len(record.Key))
			}
		})
	})

	t.Run("PopBatch", func(t *testing.T) {
		t.Run("empty", func(t *testing.T) {
			b := batch{}
			batch := b.PopBatch(1)
			assert.Empty(t, batch)
		})

		t.Run("not multiple", func(t *testing.T) {
			b := batch{}
			for i := 0; i < 31; i++ {
				require.NoError(t, b.Add(&Audit{}))
			}
			for i := 0; i < 3; i++ {
				assert.Len(t, b.PopBatch(10), 10)
			}
			assert.Len(t, b.PopBatch(10), 1)
			assert.Len(t, b.PopBatch(10), 0)
		})

		t.Run("a multiple", func(t *testing.T) {
			b := batch{}
			for i := 0; i < 30; i++ {
				require.NoError(t, b.Add(&Audit{}))
			}
			for i := 0; i < 3; i++ {
				assert.Len(t, b.PopBatch(10), 10)
			}
			assert.Len(t, b.PopBatch(10), 0)
		})
	})
}
