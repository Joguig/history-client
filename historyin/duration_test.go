package historyin

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDuration(t *testing.T) {
	tD := time.Duration(69) * time.Second

	marshalled, err := json.Marshal(Duration(tD))
	require.NoError(t, err)

	var unMarshalled Duration
	require.NoError(t, json.Unmarshal(marshalled, &unMarshalled))
	assert.Equal(t, tD, time.Duration(unMarshalled))
}
