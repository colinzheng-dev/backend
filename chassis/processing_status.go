
package chassis

import (
"database/sql/driver"
"encoding/json"

"github.com/pkg/errors"
)

// ProcessingStatus is an enumerated type representing the status of an task to be processed
type ProcessingStatus uint

const (
	// Pending means that the task is due.
	Pending = iota
	// Processing means that the task is being processed.
	Processing
	// Completed means that the task finished without any issue.
	Completed
	// Error means that the task finished with an error
	Error

)

// String converts a processing status to its string representation.
func (ps ProcessingStatus) String() string {
	switch ps {
	case Pending:
		return "pending"
	case Processing:
		return "processing"
	case Completed:
		return "completed"
	case Error:
		return "error"
	default:
		return "<unknown status>"
	}
}

// FromString does checked conversion from a string to a ProcessingStatus.
func (ps *ProcessingStatus) FromString(s string) error {
	switch s {
	case "pending":
		*ps = Pending
	case "processing":
		*ps = Processing
	case "completed":
		*ps = Completed
	case "error":
		*ps = Error
	default:
		return errors.New("unknown processing status '" + s + "'")
	}
	return nil
}

// MarshalJSON converts an internal processing status to JSON.
func (ps ProcessingStatus) MarshalJSON() ([]byte, error) {
	s := ps.String()
	if s == "<unknown status>" {
		return nil, errors.New("unknown processing status")
	}
	return json.Marshal(s)
}

// UnmarshalJSON unmarshals a processing status from a JSON string.
func (ps *ProcessingStatus) UnmarshalJSON(d []byte) error {
	var s string
	if err := json.Unmarshal(d, &s); err != nil {
		return errors.Wrap(err, "can't unmarshal processing status")
	}
	return ps.FromString(s)
}

// Scan implements the sql.Scanner interface.
func (ps *ProcessingStatus) Scan(src interface{}) error {
	var s string
	switch src := src.(type) {
	case string:
		s = src
	case []byte:
		s = string(src)
	default:
		return errors.New("incompatible type for ProcessingStatus")
	}
	return ps.FromString(s)
}

// Value implements the driver.Value interface.
func (ps ProcessingStatus) Value() (driver.Value, error) {
	return ps.String(), nil
}
