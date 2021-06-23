package types

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/jmoiron/sqlx/types"
)

// OtherInfo is a map of other information for payment methods.
type OtherInfo map[string]interface{}

// Scan implements the sql.Scanner interface.
func (m *OtherInfo) Scan(src interface{}) error {
	j := types.JSONText{}
	err := j.Scan(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(j, m)
}

// Value implements the driver.Value interface.
func (m OtherInfo) Value() (driver.Value, error) {
	v, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return types.JSONText(v).Value()
}