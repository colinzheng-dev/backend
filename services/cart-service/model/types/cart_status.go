package types

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

// CartStatus is an enumerated type representing the status of a cart.
type CartStatus uint

const (
	// Active means that the cart is the most recent one of a certain user.
	Active = iota
	// Completed means that the cart generated a purchased and order.
	Complete
	// Abandoned means that the user discarded the cart (clear it) to start a new one
	Abandoned
	//NotLoggedIn is a cart that the user built but not logged in to complete de purchase
	NotLoggedIn
)

// String converts a cart status to its string representation.
func (o CartStatus) String() string {
	switch o {
	case Active:
		return "active"
	case Complete:
		return "complete"
	case Abandoned:
		return "abandoned"
	case NotLoggedIn:
		return "not logged in"
	default:
		return "<unknown status>"
	}
}

// FromString does checked conversion from a string to a CartStatus.
func (o *CartStatus) FromString(s string) error {
	switch s {
	case "active":
		*o = Active
	case "complete":
		*o = Complete
	case "abandoned":
		*o = Abandoned
	case "not logged in":
		*o = NotLoggedIn
	default:
		return errors.New("unknown cart status '" + s + "'")
	}
	return nil
}

// MarshalJSON converts an internal ownership status to JSON.
func (o CartStatus) MarshalJSON() ([]byte, error) {
	s := o.String()
	if s == "<unknown status>" {
		return nil, errors.New("unknown ownership status")
	}
	return json.Marshal(s)
}

// UnmarshalJSON unmarshals an ownership status from a JSON string.
func (o *CartStatus) UnmarshalJSON(d []byte) error {
	var s string
	if err := json.Unmarshal(d, &s); err != nil {
		return errors.Wrap(err, "can't unmarshal cart status")
	}
	return o.FromString(s)
}

// Scan implements the sql.Scanner interface.
func (o *CartStatus) Scan(src interface{}) error {
	var s string
	switch src := src.(type) {
	case string:
		s = src
	case []byte:
		s = string(src)
	default:
		return errors.New("incompatible type for CartStatus")
	}
	return o.FromString(s)
}

// Value implements the driver.Value interface.
func (o CartStatus) Value() (driver.Value, error) {
	return o.String(), nil
}
