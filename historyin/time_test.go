package historyin

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTime(t *testing.T) {
	t.Run("should strip time zone when marshaled", func(t *testing.T) {
		og := Time(time.Now().
			Truncate(time.Nanosecond).
			In(time.FixedZone("Beijing Time", int((8 * time.Hour).Seconds()))))
		ogUTC := Time(time.Time(og).UTC())

		t.Run("as dynamo map", func(t *testing.T) {
			marshalled, err := dynamodbattribute.Marshal(og)
			require.NoError(t, err)

			var unmarshalled Time
			require.NoError(t, dynamodbattribute.Unmarshal(marshalled, &unmarshalled))

			assert.Equal(t, ogUTC, unmarshalled)
		})

		t.Run("as json", func(t *testing.T) {
			marshalled, err := json.Marshal(og)
			require.NoError(t, err)

			var unmarshalled Time
			require.NoError(t, json.Unmarshal(marshalled, &unmarshalled))

			assert.Equal(t, ogUTC, unmarshalled)
		})

		t.Run("unmarshal from various formats", func(t *testing.T) {
			testTime := time.Date(2009, time.November, 10, 23, 45, 27, 123456789, time.UTC)
			for _, tc := range []struct {
				Value     string
				Precision time.Duration
			}{
				{"2009-11-10T23:45:27.123456789Z", time.Nanosecond},
				{"2009-11-10T23:45:27.123456Z", time.Microsecond},
				{"2009-11-10T23:45:27.123Z", time.Millisecond},
				{"2009-11-10T23:45:27Z", time.Second},
			} {
				var unmarshalled Time
				require.NoError(t, json.Unmarshal([]byte("\""+tc.Value+"\""), &unmarshalled))
				assert.Equal(t, Time(testTime.Truncate(tc.Precision)), unmarshalled)
			}
		})

		t.Run("unmarshal zero time should be fine", func(t *testing.T) {
			var unmarshalled Time
			require.NoError(t, json.Unmarshal([]byte("\"0001-01-01T00:00:00Z\""), &unmarshalled))
			assert.True(t, time.Time(unmarshalled).IsZero())
		})

		t.Run("invalid time - missing tz", func(t *testing.T) {
			var unmarshalled Time
			require.Error(t, json.Unmarshal([]byte("\"2009-11-10T23:45:27\""), &unmarshalled))
			assert.True(t, time.Time(unmarshalled).IsZero())
		})
	})
}
