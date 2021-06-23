
package types

import (
	"database/sql/driver"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

// ThreadStatus is an enumeration that represents all the known types of thread_status.
type ThreadStatus int

// Constants for all thread_status.
const (
	Unknown ThreadStatus = iota
	Open
	Closed
	Archived
	Deleted
)

// UnmarshalJSON unmarshals a thread_status from a JSON string.
func (ts *ThreadStatus) UnmarshalJSON(d []byte) error {
	var s string
	if err := json.Unmarshal(d, &s); err != nil {
		return errors.Wrap(err, "can't unmarshal thread status")
	}
	return ts.FromString(s)
}

// FromString converts a string to an thread status.
func (ts *ThreadStatus) FromString(s string) error {
	switch strings.ToLower(s) {
	default:
		return errors.New("unknown item type '" + s + "'")
	case "open":
		*ts = Open
	case "closed":
		*ts = Closed
	case "comment":
		*ts = Archived
	case "deleted":
		*ts = Deleted
	}
	return nil
}

// String converts a post type from its internal representation to a string.
func (ts ThreadStatus) String() string {
	switch ts {
	default:
		return "<unknown thread status>"
	case Open:
		return "open"
	case Closed:
		return "closed"
	case Archived:
		return "archived"
	case Deleted:
		return "deleted"
	}
}

// MarshalJSON converts an internal post type to JSON.
func (ts ThreadStatus) MarshalJSON() ([]byte, error) {
	s := ts.String()
	if s == "<unknown thread status>" {
		return nil, errors.New("unknown thread status")
	}
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface.
func (ts *ThreadStatus) Scan(src interface{}) error {
	var s string
	switch src := src.(type) {
	case string:
		s = src
	case []byte:
		s = string(src)
	default:
		return errors.New("incompatible type for thread status")
	}
	return ts.FromString(s)
}

// Value implements the driver.Value interface.
func (ts ThreadStatus) Value() (driver.Value, error) {
	return ts.String(), nil
}