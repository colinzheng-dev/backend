package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	itemModel "github.com/veganbase/backend/services/item-service/model"
)

// PurchaseItems struct represents the data in the JSON/JSONB items column of purchase database.
// We can use struct tags to control how each field is encoded.
type PurchaseItems []PurchaseItem
type FullPurchaseItems []FullPurchaseItem

// PurchaseItem represents one entry on the array of items stored on column items in table purchases and orders
type PurchaseItem struct {
	ItemId           string  `json:"item_id"`
	ProductType      string  `json:"product_type,omitempty"`
	ItemOwner        string  `json:"item_owner,omitempty"`
	Quantity         int     `json:"quantity"`
	Price            int     `json:"price"`
	Currency         string  `json:"currency"`
	UniqueIdentifier string  `json:"unique_identifier"`
	OtherInfo        InfoMap `json:"other_info,omitempty"`
}

// Make the PurchaseItems struct implement the driver.Valuer interface. This method
// simply returns the JSON-encoded representation of the struct.
func (pi PurchaseItems) Value() (driver.Value, error) {
	return json.Marshal(pi)
}

func (fpi FullPurchaseItems) Value() (driver.Value, error) {
	return json.Marshal(fpi)
}

// Make the PurchaseItems struct implement the sql.Scanner interface. This method
// simply decodes a JSON-encoded value into the struct fields.
func (pi *PurchaseItems) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &pi)
}

// PurchaseItem represents one entry on the array of items stored on column items in table purchases and orders
type FullPurchaseItem struct {
	Item        itemModel.Info `json:"item"`
	ProductType string         `json:"product_type,omitempty"`
	ItemOwner   string         `json:"item_owner,omitempty"`
	Quantity    int            `json:"quantity"`
	Price       int            `json:"price"`
	Currency    string         `json:"currency"`
	OtherInfo   InfoMap        `json:"other_info,omitempty"`
}

func GetFullPurchaseItem(rawPurItem *PurchaseItem, itemInfo *itemModel.Info) *FullPurchaseItem {
	view := FullPurchaseItem{}
	view.Item = *itemInfo
	view.ProductType = rawPurItem.ProductType
	view.Quantity = rawPurItem.Quantity
	view.Price = rawPurItem.Price
	view.Currency = rawPurItem.Currency
	view.OtherInfo = rawPurItem.OtherInfo
	view.ItemOwner = rawPurItem.ItemOwner
	return &view
}
