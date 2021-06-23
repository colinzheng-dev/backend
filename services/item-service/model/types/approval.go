package types

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

// ApprovalState is an enumerated type representing the progress of an
// item from its initial unapproved state to the approved state where
// it's visible to all users.
type ApprovalState uint

const (
	// Pending represents an item after its initial creation: it is not
	// yet visible to users apart from administrators and its owner, and
	// must be approved by an administrator.
	Pending = iota

	// Approved represents an item that has been approved for release by
	// an administrator, and is now visible to all users.
	Approved

	// Rejected represents an article that an administrator rejected,
	// and so it not visible to any users apart from its owner and
	// administrators.
	Rejected
)

// String converts an approval state to its string representation.
func (a ApprovalState) String() string {
	switch a {
	case Pending:
		return "pending"
	case Approved:
		return "approved"
	case Rejected:
		return "rejected"
	default:
		return "<unknown approval state>"
	}
}

// FromString does checked conversion from a string to an
// ApprovalState.
func (a *ApprovalState) FromString(s string) error {
	switch s {
	case "pending":
		*a = Pending
	case "approved":
		*a = Approved
	case "rejected":
		*a = Rejected
	default:
		return errors.New("unknown approval state '" + s + "'")
	}
	return nil
}

// MarshalJSON converts an internal approval status to JSON.
func (a ApprovalState) MarshalJSON() ([]byte, error) {
	s := a.String()
	if s == "<unknown approval state>" {
		return nil, errors.New("unknown approval state")
	}
	return json.Marshal(s)
}

// UnmarshalJSON unmarshals an approval state from a JSON string.
func (a *ApprovalState) UnmarshalJSON(d []byte) error {
	var s string
	if err := json.Unmarshal(d, &s); err != nil {
		return errors.Wrap(err, "can't unmarshal approval state")
	}
	return a.FromString(s)
}

// Scan implements the sql.Scanner interface.
func (a *ApprovalState) Scan(src interface{}) error {
	var s string
	switch src := src.(type) {
	case string:
		s = src
	case []byte:
		s = string(src)
	default:
		return errors.New("incompatible type for ApprovalState")
	}
	return a.FromString(s)
}

// Value implements the driver.Value interface.
func (a ApprovalState) Value() (driver.Value, error) {
	return a.String(), nil
}
