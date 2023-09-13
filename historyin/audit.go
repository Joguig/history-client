package historyin

import (
	"bytes"
	"encoding/json"
	"strconv"
	"time"

	uuid "github.com/satori/go.uuid"
)

// defaultTTL of 5 years
const defaultTTL = time.Duration(5*365*24*3600) * time.Second

// this is a v4 uuid used to generate a v5 uuid.
// this string is shared with the history-client-ruby repo
const uuidNamespaceStr = "d45db28b-34ed-4738-8f7a-db2a993feadb"

// Audit is an auditable event
type Audit struct {
	// Required Attributes
	Action       string
	UserType     string
	UserID       string
	ResourceType string
	ResourceID   string
	Description  string
	Changes      []ChangeSet

	// Optional Attributes
	UUID      UUID
	CreatedAt Time
	TTL       Duration
}

// ExpiredAt returns the time when audit will be expired
func (a *Audit) ExpiredAt() Time {
	return Time(time.Time(a.CreatedAt).Add(time.Duration(a.TTL)))
}

// MarshalJSON implements json.Marshaller
func (a *Audit) MarshalJSON() ([]byte, error) {
	return json.Marshal(audit{
		UUID:         a.UUID,
		Action:       a.Action,
		UserType:     a.UserType,
		UserID:       a.UserID,
		ResourceType: a.ResourceType,
		ResourceID:   a.ResourceID,
		Description:  a.Description,
		CreatedAt:    a.CreatedAt,
		ExpiredAt:    a.ExpiredAt(),
		Expiry:       a.TTL,
		Changes:      a.Changes,
	})
}

// UnmarshalJSON implements json.Unmarshaller
func (a *Audit) UnmarshalJSON(data []byte) error {
	var raw audit
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	a.UUID = raw.UUID
	a.Action = raw.Action
	a.UserType = raw.UserType
	a.UserID = raw.UserID
	a.ResourceType = raw.ResourceType
	a.ResourceID = raw.ResourceID
	a.Description = raw.Description
	a.CreatedAt = raw.CreatedAt
	a.TTL = raw.Expiry
	a.Changes = raw.Changes

	return nil
}

// fillOptional fills fields that are marked as optional if not set
func (a *Audit) fillOptional() error {
	if a.UUID == "" {
		if err := a.fillUUID(); err != nil {
			return err
		}
	}

	if time.Duration(a.TTL) == time.Duration(0) {
		a.fillTTL()
	}

	if time.Time(a.CreatedAt).IsZero() {
		a.fillCreatedAt()
	}
	return nil
}

func (a *Audit) fillUUID() (err error) {
	namespace, err := a.uuidNamespace()
	if err != nil {
		return err
	}
	a.UUID = UUID(uuid.NewV5(namespace, a.uuidName()).String())
	return
}

func (a *Audit) fillTTL() {
	a.TTL = Duration(defaultTTL)
}

func (a *Audit) fillCreatedAt() {
	a.CreatedAt = Time(time.Now())
}

func (a *Audit) uuidNamespace() (uuid.UUID, error) {
	return uuid.FromString(uuidNamespaceStr)
}

func (a *Audit) uuidName() string {
	var buffer bytes.Buffer

	buffer.WriteString(a.Action)
	buffer.WriteString(a.UserType)
	buffer.WriteString(a.UserID)
	buffer.WriteString(a.ResourceType)
	buffer.WriteString(a.ResourceID)
	buffer.WriteString(a.Description)
	buffer.WriteString(time.Time(a.CreatedAt).String())
	buffer.WriteString(strconv.FormatInt(int64(time.Duration(a.TTL).Seconds()), 10))

	return buffer.String()
}

// json serializable struct to be written
type audit struct {
	UUID         UUID        `json:"uuid"`
	Action       string      `json:"action"`
	UserType     string      `json:"user_type"`
	UserID       string      `json:"user_id"`
	ResourceType string      `json:"resource_type"`
	ResourceID   string      `json:"resource_id"`
	Description  string      `json:"description"`
	CreatedAt    Time        `json:"created_at"`
	ExpiredAt    Time        `json:"expired_at,omitempty"`
	Expiry       Duration    `json:"expiry,omitempty"`
	Changes      []ChangeSet `json:"changes"`
}

// ChangeSet is a change of an attribute
type ChangeSet struct {
	Attribute string `json:"attribute"`
	OldValue  string `json:"old_value"`
	NewValue  string `json:"new_value"`
}
