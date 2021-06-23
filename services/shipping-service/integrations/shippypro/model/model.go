package model

type Method string

const (
	PutOrder Method = "PutOrder"
)

type Order struct {
	Method Method `json:"Method"`
	Params Params `json:"Params"`
}

type Params struct {
	ToAddress          ToAddress `json:"to_address"`
	Parcels            []Parcel  `json:"parcels"`
	Items              []Item    `json:"items"`
	TransactionID      string    `json:"TransactionID"`
	Date               int64     `json:"Date"`
	Currency           string    `json:"Currency"`
	ItemsCount         int64     `json:"ItemsCount"`
	ContentDescription string    `json:"ContentDescription"`
	Total              float64   `json:"Total"`
	Status             string    `json:"Status"`
	APIOrdersID        int64     `json:"APIOrdersID"`
	ShipmentAmountPaid float64   `json:"ShipmentAmountPaid"`
	Incoterm           string    `json:"Incoterm"`
	BillAccountNumber  string    `json:"BillAccountNumber"`
	PaymentMethod      string    `json:"PaymentMethod"`
	ShippingService    string    `json:"ShippingService"`
	Note               string    `json:"Note"`
}

type Item struct {
	Title    string  `json:"title"`
	Imageurl string  `json:"imageurl"`
	Quantity int64   `json:"quantity"`
	Price    float64 `json:"price"`
	Sku      string  `json:"sku"`
}

type Parcel struct {
	Length int64 `json:"length"`
	Width  int64 `json:"width"`
	Height int64 `json:"height"`
	Weight int64 `json:"weight"`
}

type ToAddress struct {
	Name    string `json:"name"`
	Company string `json:"company"`
	Street1 string `json:"street1"`
	Street2 string `json:"street2"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
	Country string `json:"country"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
}
