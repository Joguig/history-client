package historyin

import (
	"encoding/json"
	"time"
)

// Duration is a json serializable duration
type Duration time.Duration

// MarshalJSON implements json.Marshaller
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(time.Duration(d).Seconds()))
}

// UnmarshalJSON implements json.Unmarshaller
func (d *Duration) UnmarshalJSON(data []byte) error {
	var asInt64 int64
	err := json.Unmarshal(data, &asInt64)
	if err != nil {
		return err
	}
	*d = Duration(time.Duration(asInt64) * time.Second)
	return nil
}
