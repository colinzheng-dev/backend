package types

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/jmoiron/sqlx/types"
)

type OtherInfo map[string]interface{}

// Scan implements the sql.Scanner interface.
func (ai *OtherInfo) Scan(src interface{}) error {
	j := types.JSONText{}
	err := j.Scan(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(j, ai)
}

// Value implements the driver.Value interface.
func (ai OtherInfo) Value() (driver.Value, error) {
	v, err := json.Marshal(ai)
	if err != nil {
		return nil, err
	}
	return types.JSONText(v).Value()
}
