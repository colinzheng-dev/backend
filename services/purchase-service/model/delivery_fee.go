package model

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/jmoiron/sqlx/types"
)
type DeliveryFees []DeliveryFee

type DeliveryFee struct {
	Seller   string `json:"seller"`
	Price    int    `json:"price"`
	Currency string `json:"currency"`
}

// Make the DeliveryFee struct implement the driver.Valuer interface. This method
// simply returns the JSON-encoded representation of the struct.
func (df *DeliveryFee) Value() (driver.Value, error) {
	v, err := json.Marshal(df)
	if err != nil {
		return nil, err
	}
	return types.JSONText(v).Value()
}
// Make the DeliveryFees struct implement the driver.Valuer interface. This method
// simply returns the JSON-encoded representation of the struct.
func (dfs *DeliveryFees) Value() (driver.Value, error) {
	v, err := json.Marshal(dfs)
	if err != nil {
		return nil, err
	}
	return types.JSONText(v).Value()
}

// Make the DeliveryFee struct implement the sql.Scanner interface. This method
// simply decodes a JSON-encoded value into the struct fields.
func (df *DeliveryFee) Scan(src interface{}) error {
	j := types.JSONText{}
	err := j.Scan(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(j, df)
}

// Make the DeliveryFees struct implement the sql.Scanner interface. This method
// simply decodes a JSON-encoded value into the struct fields.

func (dfs *DeliveryFees) Scan(src interface{}) error {
	j := types.JSONText{}
	err := j.Scan(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(j, dfs)
}