package model

type ItemTypeInfo struct {
	ItemType string `json:"item_type,omitempty" db:"item_type"`
	Quantity int    `json:"quantity,omitempty" db:"quantity"`
}
