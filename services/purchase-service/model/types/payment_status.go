package types

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

// PaymentStatus is an enumerated type representing the status of a payment.
type PaymentStatus uint

// String converts a payment status to its string representation.
func (p PaymentStatus) String() string {
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
func (p *PaymentStatus) FromString(s string) error {
	switch s {
	case "pending":
		*p = Pending
	case "failed":
		*p = Failed
	case "completed":
		*p = Completed
	default:
		return errors.New("unknown payment status '" + s + "'")
	}
	return nil
}

// MarshalJSON converts an internal payment status to JSON.
func (p PaymentStatus) MarshalJSON() ([]byte, error) {
	s := p.String()
	if s == "<unknown status>" {
		return nil, errors.New("unknown payment status")
	}
	return json.Marshal(s)
}

// UnmarshalJSON unmarshals a payment status from a JSON string.
func (p *PaymentStatus) UnmarshalJSON(d []byte) error {
	var s string
	if err := json.Unmarshal(d, &s); err != nil {
		return errors.Wrap(err, "can't unmarshal payment status")
	}
	return p.FromString(s)
}

// Scan implements the sql.Scanner interface.
func (p *PaymentStatus) Scan(src interface{}) error {
	var s string
	switch src := src.(type) {
	case string:
		s = src
	case []byte:
		s = string(src)
	default:
		return errors.New("incompatible type for PaymentStatus")
	}
	return p.FromString(s)
}

// Value implements the driver.Value interface.
func (p PaymentStatus) Value() (driver.Value, error) {
	return p.String(), nil
}
