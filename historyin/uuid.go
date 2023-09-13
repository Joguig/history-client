package historyin

import (
	"encoding/json"
)

// InvalidUUIDError is the error returned for an invalid UUID
type InvalidUUIDError struct{}

func (e *InvalidUUIDError) Error() string {
	return "Invalid UUID"
}

// UUID is an event identifier
type UUID string

// MarshalJSON implements json.Marshaller
func (uuid UUID) MarshalJSON() ([]byte, error) {
	if len(uuid) != 36 {
		return nil, &InvalidUUIDError{}
	}
	return json.Marshal(string(uuid))
}

// UnmarshalJSON implements json.Marshaller
func (uuid *UUID) UnmarshalJSON(data []byte) error {
	var asStr string
	if err := json.Unmarshal(data, &asStr); err != nil {
		return err
	}

	if len(asStr) != 36 {
		return &InvalidUUIDError{}
	}

	*uuid = UUID(asStr)
	return nil
}
