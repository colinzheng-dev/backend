package types

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/jmoiron/sqlx/types"
)

// AttrMap is a map of attributes.
type AttrMap map[string]interface{}

// Scan implements the sql.Scanner interface.
func (m *AttrMap) Scan(src interface{}) error {
	j := types.JSONText{}
	err := j.Scan(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(j, m)
}

// Value implements the driver.Value interface.
func (m AttrMap) Value() (driver.Value, error) {
	v, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return types.JSONText(v).Value()
}
