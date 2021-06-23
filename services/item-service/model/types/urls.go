package types

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/jmoiron/sqlx/types"
)

//go:generate go run ../../tools/gen_url_types/main.go

// URL is a web link.
type URL string

// URLMap is a list of URLs categorised by type.
type URLMap map[URLType]string

// Scan implements the sql.Scanner interface.
func (m *URLMap) Scan(src interface{}) error {
	j := types.JSONText{}
	err := j.Scan(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(j, m)
}

// Value implements the driver.Value interface.
func (m URLMap) Value() (driver.Value, error) {
	v, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return types.JSONText(v).Value()
}
