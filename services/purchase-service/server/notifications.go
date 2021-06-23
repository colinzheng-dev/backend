package server

import (
	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis"
	item "github.com/veganbase/backend/services/item-service/model"
	"github.com/veganbase/backend/services/purchase-service/events"
	"github.com/veganbase/backend/services/purchase-service/model"
	usr "github.com/veganbase/backend/services/user-service/model"
	"strconv"
)

func (s *Server) sendPurchaseCreatedNotifications(p *model.FullPurchase) {
	//obtaining unique ids for all items and users/orgs
	userIds, itemIds := getUniqueIds(p)

	userInfo, err := s.userSvc.Info(userIds)
	if err != nil {
		log.Error().Err(err).Msg("purchase-created: could not obtain users' information from user-service")
		return
	}

	itemsInfo, err := s.itemSvc.GetItemsInfo(itemIds)
	if err != nil {
		log.Error().Err(err).Msg("purchase-created: could not obtain items information from item-service")
		return
	}

	//SENDING purchase-created notification to buyer
	notification := buildPurchaseCreatedNotificationMsg(p, itemsInfo, userInfo[p.BuyerID])
	if err = chassis.Emit(s, events.PurchaseCreatedTopic, notification); err != nil {
		log.Error().Err(err).Msg("purchase-created: could not send event")
	}

	//SENDING order-created notification to sellers
	for _, o := range *p.Orders {
		notification := buildOrderCreatedNotificationMsg(o, itemsInfo, userInfo[o.Seller])
		if err = chassis.Emit(s, events.OrderCreatedTopic, notification); err != nil {
			log.Error().Err(err).Msg("order-created: could not send event")
		}
	}

	//SENDING booking-created notification to hosts
	for _, b := range *p.Bookings {
		notification := buildBookingCreatedNotificationMsg(b, itemsInfo, userInfo[b.Host])
		if err = chassis.Emit(s, events.BookingCreatedTopic, notification); err != nil {
			log.Error().Err(err).Msg("booking-created: could not send event")
		}
	}
	return
}

// extracts all unique ids for users/orgs and items
func getUniqueIds(p *model.FullPurchase) ([]string, []string) {
	userIds := []string{}
	itemIds := []string{}
	uniqueUserIds := map[string]bool{}
	uniqueItemIds := map[string]bool{}
	for _, it := range p.Items {
		uniqueUserIds[it.ItemOwner] = true
		uniqueItemIds[it.ItemId] = true
	}
	uniqueUserIds[p.BuyerID] = true

	for k, _ := range uniqueUserIds {
		userIds = append(userIds, k)
	}

	for k, _ := range uniqueItemIds {
		itemIds = append(itemIds, k)
	}

	return userIds, itemIds
}

func buildBookingCreatedNotificationMsg(b model.Booking, itemsInfo map[string]*item.Info, hostInfo *usr.Info) *chassis.GenericEmailMsg {
	data := chassis.GenericMap{}
	//booking information
	data["host"] = hostInfo.Name
	data["booking_id"] = b.Id

	//item information
	it := chassis.GenericMap{}
	it["item_id"] = b.ItemID
	it["name"] = itemsInfo[b.ItemID].Name
	it["slug"] = itemsInfo[b.ItemID].Slug
	//bInfo["slug"] = itemsInfo[b.ItemID].Image TODO: MAYBE SEND ITEM IMG
	it["currency"] = b.BookingInfo.Currency
	it["price"] = b.BookingInfo.Price
	it["formatted_price"] = chassis.FormatCurrencyValue(b.BookingInfo.Currency, b.BookingInfo.Price*b.BookingInfo.Quantity)
	it["quantity"] = strconv.Itoa(b.BookingInfo.Quantity)
	data["item"] = it

	//booking additional information
	if guests, ok := b.BookingInfo.OtherInfo["guests"]; ok {
		guestsMap := guests.(map[string]interface{})
		data["has_guests"] = "true"
		data["adults"] = int(guestsMap["adults"].(float64))
		data["children"] = int(guestsMap["children"].(float64))
		data["infants"] = int(guestsMap["infants"].(float64))
	} else {
		data["has_guests"] = "false"
	}

	if period, ok := b.BookingInfo.OtherInfo["period"]; ok {
		periodMap := period.(map[string]interface{})
		data["has_period"] = "true"
		data["start"] = periodMap["start"].(string)
		data["end"] = periodMap["end"].(string)
	} else {
		data["has_period"] = "false"
	}

	if timeStart, ok := b.BookingInfo.OtherInfo["time_start"]; ok {
		data["has_time_start"] = "true"
		data["time_start"] = timeStart.(string)
	} else {
		data["has_time_start"] = "false"
	}

	msg := chassis.GenericEmailMsg{
		FixedFields: chassis.FixedFields{
			Site:     "ethical.id",
			Language: "en",
			Email:    *hostInfo.Email,
		},
		Data: data,
	}

	return &msg
}

func buildOrderCreatedNotificationMsg(o model.Order, itemsInfo map[string]*item.Info, sellerInfo *usr.Info) *chassis.GenericEmailMsg {
	data := chassis.GenericMap{}

	data["order_id"] = o.Id
	data["seller"] = sellerInfo.Name
	data["qty_items"] = len(o.Items)

	var total int
	var currency string
	var i int

	for _, it := range o.Items {
		i++
		itemAsMap := make(map[string]interface{})
		itemAsMap["item_id"] = it.ItemId
		itemAsMap["name"] = itemsInfo[it.ItemId].Name
		itemAsMap["slug"] = itemsInfo[it.ItemId].Slug
		//itemAsMap["image"] = itemsInfo[item.ItemId].Image TODO: MAYBE SEND ITEM IMG
		itemAsMap["price"] = strconv.Itoa(it.Price)
		itemAsMap["currency"] = it.Currency
		itemAsMap["formatted_price"] = chassis.FormatCurrencyValue(it.Currency, it.Price)
		itemAsMap["quantity"] = strconv.Itoa(it.Quantity)

		data["item"+strconv.Itoa(i)] = itemAsMap

		total += it.Price * it.Quantity
		currency = it.Currency
	}

	if o.DeliveryFee != nil && o.DeliveryFee.Price > 0 {
		total += o.DeliveryFee.Price
		data["delivery_cost"] = o.DeliveryFee.Price
		data["delivery_currency"] = o.DeliveryFee.Currency
		data["formatted_delivery_cost"] = chassis.FormatCurrencyValue(currency, o.DeliveryFee.Price)
	} else {
		data["formatted_delivery_cost"] = "FREE"
	}

	data["total"] = chassis.FormatCurrencyValue(currency, total)

	msg := chassis.GenericEmailMsg{
		FixedFields: chassis.FixedFields{
			Site:     "ethical.id",
			Language: "en",
			Email:    *sellerInfo.Email,
		},
		Data: data,
	}

	return &msg
}

func buildPurchaseCreatedNotificationMsg(p *model.FullPurchase, itemsInfo map[string]*item.Info, info *usr.Info) *chassis.GenericEmailMsg {

	data := chassis.GenericMap{}

	data["purchase_id"] = p.Id
	data["customer_name"] = info.Name
	data["qty_orders"] = len(*p.Orders)
	data["qty_bookings"] = len(*p.Bookings)

	var total int

	var currency string
	orders := []chassis.GenericMap{}
	for _, o := range *p.Orders {
		order := chassis.GenericMap{}
		items := []chassis.GenericMap{}
		order["order_id"] = o.Id
		order["seller"] = o.Seller
		order["qty_items"] = len(o.Items)

		for _, it := range o.Items {
			currency = it.Currency
			total += it.Price * it.Quantity

			itemAsMap := make(map[string]interface{})
			itemAsMap["item_id"] = it.ItemId
			itemAsMap["name"] = itemsInfo[it.ItemId].Name
			itemAsMap["slug"] = itemsInfo[it.ItemId].Slug
			//itemAsMap["image"] = itemsInfo[item.ItemId].Image TODO: MAYBE SEND ITEM IMG
			itemAsMap["price"] = strconv.Itoa(it.Price)
			itemAsMap["currency"] = it.Currency
			itemAsMap["formatted_price"] = chassis.FormatCurrencyValue(it.Currency, it.Price)
			itemAsMap["quantity"] = strconv.Itoa(it.Quantity)

			items = append(items, itemAsMap)
		}

		if o.DeliveryFee != nil && o.DeliveryFee.Price > 0 {
			order["delivery_cost"] = o.DeliveryFee.Price
			order["delivery_currency"] = o.DeliveryFee.Currency
			order["formatted_delivery_cost"] = chassis.FormatCurrencyValue(currency, o.DeliveryFee.Price)
		} else {
			order["formatted_delivery_cost"] = "FREE"
		}
		orders = append(orders, order)
	}
	data["orders"] = orders
	data["display_summary"] = "false"

	bookings := []chassis.GenericMap{}
	for _, b := range *p.Bookings {
		//item information
		booking := chassis.GenericMap{}
		booking["item_id"] = b.ItemID
		booking["name"] = itemsInfo[b.ItemID].Name
		booking["slug"] = itemsInfo[b.ItemID].Slug
		//bInfo["slug"] = itemsInfo[b.ItemID].Image TODO: MAYBE SEND ITEM IMG
		booking["currency"] = b.BookingInfo.Currency
		booking["price"] = b.BookingInfo.Price
		booking["formatted_price"] = chassis.FormatCurrencyValue(b.BookingInfo.Currency, b.BookingInfo.Price*b.BookingInfo.Quantity)
		booking["quantity"] = strconv.Itoa(b.BookingInfo.Quantity)


		//booking additional information
		if guests, ok := b.BookingInfo.OtherInfo["guests"]; ok {
			guestsMap := guests.(map[string]interface{})
			booking["has_guests"] = "true"
			booking["adults"] = int(guestsMap["adults"].(float64))
			booking["children"] = int(guestsMap["children"].(float64))
			booking["infants"] = int(guestsMap["infants"].(float64))
		} else {
			booking["has_guests"] = "false"
		}

		if period, ok := b.BookingInfo.OtherInfo["period"]; ok {
			periodMap := period.(map[string]interface{})
			booking["has_period"] = "true"
			booking["start"] = periodMap["start"].(string)
			booking["end"] = periodMap["end"].(string)
		} else {
			booking["has_period"] = "false"
		}

		if timeStart, ok := b.BookingInfo.OtherInfo["time_start"]; ok {
			booking["has_time_start"] = "true"
			booking["time_start"] = timeStart.(string)
		} else {
			booking["has_time_start"] = "false"
		}
		bookings = append(bookings, booking)
	}

	data["bookings"] = bookings
	msg := chassis.GenericEmailMsg{
		FixedFields: chassis.FixedFields{
			Site:     "veganbase",
			Language: "en",
			Email:    *info.Email,
		},
		Data: data,
	}

	return &msg
}
