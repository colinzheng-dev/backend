package types

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

// SubscriptionStatus is an enumerated type representing the status of an item subscription.
type SubscriptionStatus uint

const (
	// Active means that the subscription item is active and will be delivered based on its delivery schedule.
	Active = iota
	// Paused means that the user paused the subscriptions.
	Paused
	// Deleted means that the subscription status was deleted.
	Deleted

)

// String converts a subscription status to its string representation.
func (ss SubscriptionStatus) String() string {
	switch ss {
	case Active:
		return "active"
	case Paused:
		return "paused"
	case Deleted:
		return "deleted"
	default:
		return "<unknown status>"
	}
}

// FromString does checked conversion from a string to a PurchaseStatus.
func (ss *SubscriptionStatus) FromString(s string) error {
	switch s {
	case "active":
		*ss = Active
	case "paused":
		*ss = Paused
	case "deleted":
		*ss = Deleted
	default:
		return errors.New("unknown subscription status '" + s + "'")
	}
	return nil
}

// MarshalJSON converts an internal subscription status to JSON.
func (ss SubscriptionStatus) MarshalJSON() ([]byte, error) {
	s := ss.String()
	if s == "<unknown status>" {
		return nil, errors.New("unknown subscription status")
	}
	return json.Marshal(s)
}

// UnmarshalJSON unmarshals a subscription status from a JSON string.
func (ss *SubscriptionStatus) UnmarshalJSON(d []byte) error {
	var s string
	if err := json.Unmarshal(d, &s); err != nil {
		return errors.Wrap(err, "can't unmarshal cart status")
	}
	return ss.FromString(s)
}

// Scan implements the sql.Scanner interface.
func (ss *SubscriptionStatus) Scan(src interface{}) error {
	var s string
	switch src := src.(type) {
	case string:
		s = src
	case []byte:
		s = string(src)
	default:
		return errors.New("incompatible type for SubscriptionStatus")
	}
	return ss.FromString(s)
}

// Value implements the driver.Value interface.
func (ss SubscriptionStatus) Value() (driver.Value, error) {
	return ss.String(), nil
}
