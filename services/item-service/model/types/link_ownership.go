package types

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

// LinkOwnershipClass is an enumerated type representing the possible
// ownership classes of an inter-item link.
type LinkOwnershipClass uint

const (
	// OwnerToOwner represents a link type that may only be created
	// between a pair of items for which the owner of both items is the
	// creating user.
	OwnerToOwner = iota

	// OwnerToAny represents a link type that may be created by the
	// owner of the origin item.
	OwnerToAny

	// AnyToOwner represents a link type that may be created by the
	// owner of the target item.
	AnyToOwner
)

// String converts a link ownership class to its string
// representation.
func (a LinkOwnershipClass) String() string {
	switch a {
	case OwnerToOwner:
		return "owner-to-owner"
	case OwnerToAny:
		return "owner-to-any"
	case AnyToOwner:
		return "any-to-owner"
	default:
		return "<unknown link ownership class>"
	}
}

// FromString does checked conversion from a string to an
// LinkOwnershipClass.
func (a *LinkOwnershipClass) FromString(s string) error {
	switch s {
	case "owner-to-owner":
		*a = OwnerToOwner
	case "owner-to-any":
		*a = OwnerToAny
	case "any-to-owner":
		*a = AnyToOwner
	default:
		return errors.New("unknown link ownership class '" + s + "'")
	}
	return nil
}

// MarshalJSON converts an internal link ownership class to JSON.
func (a LinkOwnershipClass) MarshalJSON() ([]byte, error) {
	s := a.String()
	if s == "<unknown link ownership class>" {
		return nil, errors.New("unknown link ownership class")
	}
	return json.Marshal(s)
}

// UnmarshalJSON unmarshals a link ownership class from a JSON string.
func (a *LinkOwnershipClass) UnmarshalJSON(d []byte) error {
	var s string
	if err := json.Unmarshal(d, &s); err != nil {
		return errors.Wrap(err, "can't unmarshal link ownership class")
	}
	return a.FromString(s)
}

// Scan implements the sql.Scanner interface.
func (a *LinkOwnershipClass) Scan(src interface{}) error {
	var s string
	switch src := src.(type) {
	case string:
		s = src
	case []byte:
		s = string(src)
	default:
		return errors.New("incompatible type for LinkOwnershipClass")
	}
	return a.FromString(s)
}

// Value implements the driver.Value interface.
func (a LinkOwnershipClass) Value() (driver.Value, error) {
	return a.String(), nil
}
