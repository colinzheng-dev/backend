package events

import (
	"encoding/json"
	"github.com/veganbase/backend/chassis"
	purModel "github.com/veganbase/backend/services/purchase-service/model"
	"time"
)

// Event types for payments creation and update.
const (
	PaymentCreated = "payment-created"
	PaymentUpdated = "payment-updated"
	OrderPlaced    = "order.placed"
)

// Topics that will trigger events on other services.
const (
	PaymentStatusTopic   = "payment-status-topic"
	PaymentReceivedTopic = "payment-received-topic"
	SaleCompleteTopic    = "sale-complete-topic"
)

func BuildOrderPlacedEventPayload(ord purModel.Order) json.RawMessage {
	items := []OrderItem{}
	grams := "gr"
	for _, item := range ord.Items {
		ordItem := OrderItem{
			OfferingID:  item.ItemId,
			ProductType: item.ProductType,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Currency:    item.Currency,
		}

		chassis.StringField(&ordItem.Name, item.OtherInfo, "name")
		chassis.IntField(ordItem.Weight, item.OtherInfo, "weight")
		if ordItem.Weight != nil {
			ordItem.WeightUnit = &grams
		}
		chassis.StringField(&ordItem.OriginalItemID, item.OtherInfo, "original_item_id")
		chassis.StringField(&ordItem.SKU, item.OtherInfo, "sku")
		chassis.StringField(&ordItem.HSCode, item.OtherInfo, "hs_code")
		chassis.StringField(&ordItem.Barcode, item.OtherInfo, "barcode")

		items = append(items, ordItem)
	}

	cus := CustomerInfo{CustomerID: ord.BuyerID,}
	recipient := ord.OrderInfo["recipient"].(map[string]interface{})
	chassis.StringField(&cus.FirstName, recipient, "firstname")
	chassis.StringField(&cus.LastName, recipient, "lastname")
	chassis.StringField(&cus.ContactNumber, recipient, "contact_phone")
	chassis.StringField(&cus.ContactEmail, recipient, "contact_email")
	chassis.StringField(&cus.Company, recipient, "company")

	addr := Address{}
	address := ord.OrderInfo["address"].(map[string]interface{})
	chassis.StringField(&addr.StreetAddress, address, "street_address")
	chassis.StringField(&addr.City, address, "city")
	chassis.StringField(&addr.Postcode, address, "postcode")
	chassis.StringField(&addr.Country, address, "country")
	chassis.StringField(&addr.RegionPostalCode, address, "region_postal")
	chassis.StringField(&addr.HouseNumber, address, "house_number")

	orderPlaced := OrderPlacedPayload{
		OrderID:         ord.Id,
		PlacedAt:        ord.CreatedAt,
		Customer:        cus,
		DeliveryAddress: addr,
		Items:           items,
	}

	if ord.DeliveryFee != nil {
		orderPlaced.DeliveryFee = &ord.DeliveryFee.Price
		orderPlaced.DeliveryFeeCurrency = &ord.DeliveryFee.Currency
	}

	payload, _ := json.Marshal(orderPlaced)
	return payload
}

type OrderPlacedPayload struct {
	OrderID             string       `json:"order_id"`
	PlacedAt            time.Time    `json:"placed_at"`
	DeliveryFee         *int          `json:"delivery_fee"`
	DeliveryFeeCurrency *string       `json:"delivery_fee_currency"`
	Customer            CustomerInfo `json:"customer"`
	DeliveryAddress     Address      `json:"delivery_address"`
	Items               []OrderItem  `json:"items"`
}

type CustomerInfo struct {
	CustomerID    string `json:"customer_id"`
	FirstName     string `json:"firstname"`
	LastName      string `json:"lastname"`
	ContactEmail  string `json:"contact_email"`
	ContactNumber string `json:"contact_number"`
	Company       string `json:"company"`
}

type Address struct {
	StreetAddress    string `json:"street_address"`
	City             string `json:"city"`
	Postcode         string `json:"postcode"`
	Country          string `json:"country"`
	RegionPostalCode string `json:"region_postal"`
	HouseNumber      string `json:"house_number"`
}

type OrderItem struct {
	OfferingID     string  `json:"offering_item_id"`
	OriginalItemID string  `json:"item_id"`
	ProductType    string  `json:"product_type,omitempty"`
	Name           string  `json:"name"`
	SKU            string  `json:"sku,omitempty"`
	HSCode         string  `json:"hs_code,omitempty"`
	Barcode        string  `json:"barcode,omitempty"`
	Quantity       int     `json:"quantity"`
	Price          int     `json:"price"`
	Currency       string  `json:"currency"`
	Weight         *int    `json:"weight"`
	WeightUnit     *string `json:"weight_unit"`
}
