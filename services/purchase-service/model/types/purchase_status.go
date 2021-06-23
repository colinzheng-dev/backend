package types

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

// PurchaseStatus is an enumerated type representing the status of a purchase.
type PurchaseStatus uint

const (
	// Pending means that the purchase is not yet completed and waiting for some process to happen.
	Pending = iota
	// Failed means that the purchase failed at some process.
	Failed
	// Completed means that the purchase happened without any issue.
	Completed

)

// String converts a purchase status to its string representation.
func (p PurchaseStatus) String() string {
	switch p {
	case Pending:
		return "pending"
	case Failed:
		return "failed"
	case Completed:
		return "completed"
	default:
		return "<unknown status>"
	}
}

// FromString does checked conversion from a string to a PurchaseStatus.
func (p *PurchaseStatus) FromString(s string) error {
	switch s {
	case "pending":
		*p = Pending
	case "failed":
		*p = Failed
	case "completed":
		*p = Completed
	default:
		return errors.New("unknown purchase status '" + s + "'")
	}
	return nil
}

// MarshalJSON converts an internal purchase status to JSON.
func (p PurchaseStatus) MarshalJSON() ([]byte, error) {
	s := p.String()
	if s == "<unknown status>" {
		return nil, errors.New("unknown purchase status")
	}
	return json.Marshal(s)
}

// UnmarshalJSON unmarshals a purchase status from a JSON string.
func (p *PurchaseStatus) UnmarshalJSON(d []byte) error {
	var s string
	if err := json.Unmarshal(d, &s); err != nil {
		return errors.Wrap(err, "can't unmarshal cart status")
	}
	return p.FromString(s)
}

// Scan implements the sql.Scanner interface.
func (p *PurchaseStatus) Scan(src interface{}) error {
	var s string
	switch src := src.(type) {
	case string:
		s = src
	case []byte:
		s = string(src)
	default:
		return errors.New("incompatible type for CartStatus")
	}
	return p.FromString(s)
}

// Value implements the driver.Value interface.
func (p PurchaseStatus) Value() (driver.Value, error) {
	return p.String(), nil
}
