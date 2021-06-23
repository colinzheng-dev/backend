package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// BookingInfo struct represents the data in the JSON/JSONB items column of bookings table.
// We can use struct tags to control how each field is encoded.
type BookingInfo struct {
	Quantity  int     `json:"quantity"`
	Price     int     `json:"price"`
	Currency  string  `json:"currency"`
	OtherInfo InfoMap `json:"other_info"`
}

// Make the BookingInfo struct implement the driver.Valuer interface. This method
// simply returns the JSON-encoded representation of the struct.
func (bi BookingInfo) Value() (driver.Value, error) {
	return json.Marshal(bi)
}

// Make the BookingInfo struct implement the sql.Scanner interface. This method
// simply decodes a JSON-encoded value into the struct fields.
func (bi *BookingInfo) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &bi)
}
