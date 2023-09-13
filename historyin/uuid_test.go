package historyin

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUUID(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		myUUID := "abcdefabcdefabcdefabcdefabcdefabcdef"

		marshalled, err := json.Marshal(UUID(myUUID))
		require.NoError(t, err)

		var unMarshalled UUID
		require.NoError(t, json.Unmarshal(marshalled, &unMarshalled))
		assert.Equal(t, myUUID, string(unMarshalled))
	})
}
