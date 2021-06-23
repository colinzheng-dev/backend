package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// InfoMap struct represents the data in the JSON/JSONB items column of bookings table.
// We can use struct tags to control how each field is encoded.
type InfoMap map[string]interface{}

// Make the InfoMap struct implement the driver.Valuer interface. This method
// simply returns the JSON-encoded representation of the struct.
func (mi InfoMap) Value() (driver.Value, error) {
	return json.Marshal(mi)
}

// Make the InfoMap struct implement the sql.Scanner interface. This method
// simply decodes a JSON-encoded value into the struct fields.
func (mi *InfoMap) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &mi)
}
