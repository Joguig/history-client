package historyin

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	iso8601Nano   = "2006-01-02T15:04:05.000000000Z07:00"
	iso8601Micro  = "2006-01-02T15:04:05.000000Z07:00"
	iso8601Milli  = "2006-01-02T15:04:05.000Z07:00"
	iso8601Second = "2006-01-02T15:04:05Z07:00"
)

var (
	errInvalidTime = errors.New("invalid time")
)

// Time is a dynamo and json marshallable time.Time. Value truncated to
// nanosecond accuracy and time zone stripped.
type Time time.Time

// MarshalDynamoDBAttributeValue implements dynamodbattribute.Marshaler
func (t Time) MarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	av.SetN(strconv.FormatInt(int64(time.Time(t).UnixNano()), 10))
	return nil
}

// UnmarshalDynamoDBAttributeValue implements dynamodbattribute.Unmarshaler
func (t *Time) UnmarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	if av.N == nil {
		return errInvalidTime
	}

	raw, err := strconv.ParseInt(aws.StringValue(av.N), 10, 64)
	if err != nil {
		return err
	}

	*t = Time(time.Unix(0, raw).UTC())

	return nil
}

// MarshalJSON implements json.Marshaller
func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(t).Format(iso8601Nano))
}

// UnmarshalJSON implements json.Unmarshaller
func (t *Time) UnmarshalJSON(data []byte) (err error) {
	var asStr string
	if err = json.Unmarshal(data, &asStr); err != nil {
		return err
	}

	// legacy clients will send time as this if not set in their clients. we want
	// to keep this value zero if thats the case.
	if strings.HasPrefix(asStr, "0001-01-01T00:00:00Z") {
		return
	}

	var rawTime time.Time
	for _, timeFormat := range []string{
		iso8601Nano,
		iso8601Micro,
		iso8601Milli,
		iso8601Second,
	} {
		var err error
		if rawTime, err = time.Parse(timeFormat, asStr); err != nil {
			continue
		}
		break
	}

	if rawTime.IsZero() {
		return fmt.Errorf("invalid time received: %s", string(data))
	}

	*t = Time(rawTime.UTC())
	return nil
}
