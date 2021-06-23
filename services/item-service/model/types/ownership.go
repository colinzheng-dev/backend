package types

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

// OwnershipStatus is an enumerated type representing the whether an
// item is owned by its creator (the initial state) or has been
// claimed by another user (and the ownership claim has been
// approved).
type OwnershipStatus uint

const (
	// Creator represents an item after its initial creation: it is
	// owned by its creator.
	Creator = iota

	// Claimed represents an item whose ownership has been claimed by a
	// user and that ownership claim has been approved by an
	// administrator.
	Claimed
)

// String converts an ownership status to its string representation.
func (o OwnershipStatus) String() string {
	switch o {
	case Creator:
		return "creator"
	case Claimed:
		return "claimed"
	default:
		return "<unknown ownership status>"
	}
}

// FromString does checked conversion from a string to an
// OwnershipStatus.
func (o *OwnershipStatus) FromString(s string) error {
	switch s {
	case "creator":
		*o = Creator
	case "claimed":
		*o = Claimed
	default:
		return errors.New("unknown ownership status '" + s + "'")
	}
	return nil
}

// MarshalJSON converts an internal ownership status to JSON.
func (o OwnershipStatus) MarshalJSON() ([]byte, error) {
	s := o.String()
	if s == "<unknown ownership status>" {
		return nil, errors.New("unknown ownership status")
	}
	return json.Marshal(s)
}

// UnmarshalJSON unmarshals an ownership status from a JSON string.
func (o *OwnershipStatus) UnmarshalJSON(d []byte) error {
	var s string
	if err := json.Unmarshal(d, &s); err != nil {
		return errors.Wrap(err, "can't unmarshal ownership status")
	}
	return o.FromString(s)
}

// Scan implements the sql.Scanner interface.
func (o *OwnershipStatus) Scan(src interface{}) error {
	var s string
	switch src := src.(type) {
	case string:
		s = src
	case []byte:
		s = string(src)
	default:
		return errors.New("incompatible type for OwnershipStatus")
	}
	return o.FromString(s)
}

// Value implements the driver.Value interface.
func (o OwnershipStatus) Value() (driver.Value, error) {
	return o.String(), nil
}
