package server

import (
	"time"

	cartModel "github.com/veganbase/backend/services/cart-service/model"
	it "github.com/veganbase/backend/services/item-service/model"
	"github.com/veganbase/backend/services/purchase-service/model"
	"github.com/veganbase/backend/services/purchase-service/model/types"
	usr "github.com/veganbase/backend/services/user-service/model"
)

type ShortAddress struct {
	StreetAddress string `json:"street_address"`
	City          string `json:"city"`
	Postcode      string `json:"postcode"`
	Country       string `json:"country"`
	RegionPostal  string `json:"region_postal"`
	HouseNumber   string `json:"house_number"`
}

// fillPurchaseItem fills the struct that will be used inside items JSONB column on the purchases table.
func fillPurchaseItem(itemInfo it.ItemFullWithLink, cartItem cartModel.CartItemFull) types.PurchaseItem {

	purItem := types.PurchaseItem{
		ItemId:      itemInfo.ID,
		ProductType: itemInfo.ItemType.String(),
		ItemOwner:   itemInfo.Owner.ID,
		Quantity:    cartItem.Quantity,
		Price:       int(itemInfo.Attrs["price"].(float64)),
		Currency:    itemInfo.Attrs["currency"].(string),
		OtherInfo:   make(map[string]interface{}),
	}
	// try to cast to string, ignore an empty value
	purItem.UniqueIdentifier, _ = itemInfo.Attrs["unique_identifier"].(string)

	if cartItem.OtherInfo != nil {

		for k, v := range cartItem.OtherInfo {
			purItem.OtherInfo[k] = v
		}
	}

	if purItem.ProductType == it.ProductOfferingItem.String() {

		purItem.OtherInfo["original_item_id"] = itemInfo.Link.Target.ID

		if itemInfo.Attrs["unique_identifier"] != nil {
			purItem.OtherInfo["unique_identifier"] = itemInfo.Attrs["unique_identifier"]
		}
		if itemInfo.Link.Target.Attrs["sku"] != nil {
			purItem.OtherInfo["sku"] = itemInfo.Link.Target.Attrs["sku"]
		}
		if itemInfo.Link.Target.Attrs["hs_code"] != nil {
			purItem.OtherInfo["hs_code"] = itemInfo.Link.Target.Attrs["hs_code"]
		}
		if itemInfo.Link.Target.Attrs["barcode"] != nil {
			purItem.OtherInfo["barcode"] = itemInfo.Link.Target.Attrs["barcode"]
		}
		if itemInfo.Link.Target.Attrs["weight"] != nil {
			purItem.OtherInfo["weight"] = itemInfo.Link.Target.Attrs["weight"]
		}

		purItem.OtherInfo["name"] = itemInfo.Link.Target.Name
	}

	return purItem
}
func fillPurchaseItemFromSubscription(itemInfo it.ItemFullWithLink, sub model.SubscriptionItem) types.PurchaseItem {

	purItem := types.PurchaseItem{
		ItemId:      itemInfo.ID,
		ProductType: itemInfo.ItemType.String(),
		ItemOwner:   itemInfo.Owner.ID,
		Quantity:    sub.Quantity,
		Price:       int(itemInfo.Attrs["price"].(float64)),
		Currency:    itemInfo.Attrs["currency"].(string),
		OtherInfo:   make(map[string]interface{}),
	}
	if sub.OtherInfo != nil {

		for k, v := range sub.OtherInfo {
			purItem.OtherInfo[k] = v
		}
	}

	if purItem.ProductType == it.ProductOfferingItem.String() {

		purItem.OtherInfo["original_item_id"] = itemInfo.Link.Target.ID

		if itemInfo.Attrs["unique_identifier"] != nil {
			purItem.OtherInfo["unique_identifier"] = itemInfo.Attrs["unique_identifier"]
		}
		if itemInfo.Link.Target.Attrs["sku"] != nil {
			purItem.OtherInfo["sku"] = itemInfo.Link.Target.Attrs["sku"]
		}
		if itemInfo.Link.Target.Attrs["hs_code"] != nil {
			purItem.OtherInfo["hs_code"] = itemInfo.Link.Target.Attrs["hs_code"]
		}
		if itemInfo.Link.Target.Attrs["barcode"] != nil {
			purItem.OtherInfo["barcode"] = itemInfo.Link.Target.Attrs["barcode"]
		}
	}

	return purItem
}

func fillSubscriptionItem(itemInfo it.ItemFullWithLink, cartItem cartModel.CartItemFull) model.SubscriptionItem {

	subsItem := model.SubscriptionItem{
		ItemID:        itemInfo.ID,
		ItemType:      itemInfo.ItemType,
		Owner:         itemInfo.Owner.ID,
		Quantity:      cartItem.Quantity,
		DeliveryEvery: cartItem.DeliveryEvery,
		OtherInfo:     make(map[string]interface{}),
	}
	if cartItem.OtherInfo != nil {
		for k, v := range cartItem.OtherInfo {
			subsItem.OtherInfo[k] = v
		}
	}

	if subsItem.ItemType == it.ProductOfferingItem {

		subsItem.OtherInfo["original_item_id"] = itemInfo.Link.Target.ID

		if itemInfo.Attrs["unique_identifier"] != nil {
			subsItem.OtherInfo["unique_identifier"] = itemInfo.Attrs["unique_identifier"]
		}
		if itemInfo.Link.Target.Attrs["sku"] != nil {
			subsItem.OtherInfo["sku"] = itemInfo.Link.Target.Attrs["sku"]
		}
		if itemInfo.Link.Target.Attrs["hs_code"] != nil {
			subsItem.OtherInfo["hs_code"] = itemInfo.Link.Target.Attrs["hs_code"]
		}
		if itemInfo.Link.Target.Attrs["barcode"] != nil {
			subsItem.OtherInfo["barcode"] = itemInfo.Link.Target.Attrs["barcode"]
		}
	}

	subsItem.NextDelivery = (int(time.Now().Month()) + subsItem.DeliveryEvery) % 12

	return subsItem
}

//returns slices of the orders and bookings of a purchase that will be created
func fillOrdersAndBookings(buyerId string,
	itemsByOwner map[string][]types.PurchaseItem,
	deliveriesBySeller map[string]model.DeliveryFee,
	purchaseItems types.PurchaseItems, address *usr.Address) ([]model.Order, []model.Booking) {
	//creating orders
	orders := []model.Order{}
	var addr ShortAddress

	//only loop if there are items split by owner (products), because they will require address
	//the user is not requested to give address information when making a simple-purchase,
	//therefore it'll only work with bookings.
	if len(itemsByOwner) > 0 {
		addr.StreetAddress = address.StreetAddress
		addr.City = address.City
		addr.Postcode = address.Postcode
		addr.Country = address.Country
		addr.HouseNumber = address.HouseNumber
		addr.RegionPostal = address.RegionPostalCode

		moreInfo := make(map[string]interface{})
		moreInfo["address"] = addr
		moreInfo["recipient"] = address.Recipient

		for key, items := range itemsByOwner {
			var deliveryFee *model.DeliveryFee
			if delivery, ok := deliveriesBySeller[key]; ok {
				deliveryFee = &delivery
			}

			ord := model.Order{
				BuyerID:       buyerId,
				Seller:        key,
				PaymentStatus: types.Pending,
				Items:         items,
				DeliveryFee:   deliveryFee,
				OrderInfo:     moreInfo,
			}

			orders = append(orders, ord)
		}
	}

	//creating bookings
	bookings := []model.Booking{}
	for _, item := range purchaseItems {
		if item.ProductType != it.ProductOfferingItem.String() && item.ProductType != it.DishItem.String() {
			bk := model.Booking{}
			bk.BuyerID = buyerId
			bk.Host = item.ItemOwner
			bk.PaymentStatus = types.Pending
			bk.ItemID = item.ItemId
			bk.BookingInfo = types.BookingInfo{
				Quantity:  item.Quantity,
				Price:     item.Price,
				Currency:  item.Currency,
				OtherInfo: item.OtherInfo,
			}
			bookings = append(bookings, bk)
		}
	}
	return orders, bookings
}
