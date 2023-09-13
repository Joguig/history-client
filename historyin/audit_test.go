package historyin

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAudit(t *testing.T) {
	now := time.Now().UTC()
	dummyAudit := Audit{
		Action:       "my-action",
		UserType:     "my-user-type",
		UserID:       "my-user-id",
		ResourceType: "my-resource-type",
		ResourceID:   "my-resource-id",
		Description:  "my-description",
		CreatedAt:    Time(now),
		Changes: []ChangeSet{
			{"cs-attribute", "cs-old-value", "cs-new-value"},
		},
		UUID: "uuid--uuid--uuid--uuid--uuid--uuid--",
		TTL:  Duration(time.Hour),
	}

	t.Run("ExpiredAt", func(t *testing.T) {
		assert.Equal(t, Time(now.Add(time.Hour)), dummyAudit.ExpiredAt())
	})

	t.Run("Marshal", func(t *testing.T) {
		marshaled, err := json.Marshal(&dummyAudit)
		require.NoError(t, err)

		var unmarshalled Audit
		require.NoError(t, json.Unmarshal(marshaled, &unmarshalled))
		assert.Equal(t, dummyAudit, unmarshalled)
	})

	t.Run("fillOptional", func(t *testing.T) {
		t.Run("all provided", func(t *testing.T) {
			a := Audit{
				CreatedAt: Time(time.Now()),
				UUID:      "my-uuid",
				TTL:       Duration(time.Second),
			}

			filledA := new(Audit)
			*filledA = a
			require.NoError(t, filledA.fillOptional())

			assert.Equal(t, a, *filledA)
		})

		t.Run("all filled", func(t *testing.T) {
			a := &Audit{}

			require.NoError(t, a.fillOptional())

			assert.False(t, time.Time(a.CreatedAt).IsZero())
			assert.NotEmpty(t, a.TTL)
			assert.NotEmpty(t, a.UUID)
		})
	})
}
