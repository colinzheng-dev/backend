package model

import (
	itemModel "github.com/veganbase/backend/services/item-service/model"
	"github.com/veganbase/backend/services/payment-service/model"
	"github.com/veganbase/backend/services/purchase-service/model/types"
	userModel "github.com/veganbase/backend/services/user-service/model"
	"time"
)

type FullPurchase struct {
	Purchase
	Orders   *[]Order               `json:"orders,omitempty"`
	Bookings *[]Booking             `json:"bookings,omitempty"`
	Payments *[]model.PaymentIntent `json:"payments,omitempty"`
}

func FullView(purchase *Purchase, orders *[]Order, bookings *[]Booking, intents *[]model.PaymentIntent) *FullPurchase {
	view := FullPurchase{}
	view.Purchase = *purchase
	view.Orders = orders
	view.Bookings = bookings
	view.Payments = intents
	return &view
}

type FullOrder struct {
	Id            string                  `json:"id"`
	Origin        string                  `json:"origin"`
	Buyer         *userModel.Info         `json:"buyer"`
	Seller        *userModel.Info         `json:"seller"`
	PaymentStatus types.PaymentStatus     `json:"payment_status"`
	Items         types.FullPurchaseItems `json:"items"`
	DeliveryFee   *DeliveryFee            `json:"delivery_fee,omitempty"`
	OrderInfo     types.InfoMap           `json:"order_info,omitempty"`
	CreatedAt     time.Time               `json:"created_at"`
}

type FullBooking struct {
	Id            string              `json:"id"`
	Origin        string              `json:"origin"`
	Buyer         *userModel.Info     `json:"buyer"`
	Host          *userModel.Info     `json:"host"`
	Item          *itemModel.Info     `json:"item"`
	PaymentStatus types.PaymentStatus `json:"payment_status"`
	BookingInfo   types.BookingInfo   `json:"booking_info"`
	CreatedAt     time.Time           `json:"created_at"`
}

func GetFullOrder(rawOrder *Order, buyer, seller *userModel.Info, items types.FullPurchaseItems) *FullOrder {
	view := FullOrder{}
	view.Id = rawOrder.Id
	view.Origin = rawOrder.Origin
	view.Buyer = buyer
	view.Seller = seller
	view.PaymentStatus = rawOrder.PaymentStatus
	view.Items = items
	if *rawOrder.DeliveryFee != (DeliveryFee{}) {
		view.DeliveryFee = rawOrder.DeliveryFee
	}
	view.OrderInfo = rawOrder.OrderInfo
	view.CreatedAt = rawOrder.CreatedAt
	return &view
}

func GetFullBooking(rawBooking *Booking, buyer, host *userModel.Info, item *itemModel.Info) *FullBooking {
	view := FullBooking{}
	view.Id = rawBooking.Id
	view.Origin = rawBooking.Origin
	view.Buyer = buyer
	view.Host = host
	view.PaymentStatus = rawBooking.PaymentStatus
	view.Item = item
	view.BookingInfo = rawBooking.BookingInfo
	view.CreatedAt = rawBooking.CreatedAt
	return &view
}
