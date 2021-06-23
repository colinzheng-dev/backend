package server

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis"
	cartUtils "github.com/veganbase/backend/services/cart-service/server"
	it "github.com/veganbase/backend/services/item-service/model"
	"github.com/veganbase/backend/services/purchase-service/events"
	"github.com/veganbase/backend/services/purchase-service/model"
	"github.com/veganbase/backend/services/purchase-service/model/types"
	usr "github.com/veganbase/backend/services/user-service/model"

	"time"
)

const ProcessingDay = 12
const ProcessingHour = "12:00"
const DateLayout = "2006/01/02"

func (s *Server) ScheduleItemSubscriptionProcessingJobs() {
	// //TODO: may be interesting to schedule based on GMT
	// processingHour := ProcessingHour
	// if s.isDevMode {
	// 	hour, min, _ :=time.Now().UTC().Clock()
	// 	min += 1
	// 	if min >= 60 {
	// 		hour++
	// 		min = min % 60
	// 	}

	// 	if hour >= 24 {
	// 		hour = hour % 24
	// 	}
	// 	processingHour = fmt.Sprintf("%d:%d", hour, min)
	// }

	// gocron.SetLocker(&s.redisClient)
	// scheduler := gocron.NewScheduler(time.UTC)
	// _, _ = scheduler.Every(1).Day().At(processingHour).Lock().Do(s.processItemSubscriptions)
	// // scheduler starts running jobs and current thread continues to execute
	// scheduler.Start()

}

func (s *Server) processItemSubscriptions() {
	today := time.Now()
	retry := true
	for retry {
		retry = false
		ref := today.Format(DateLayout)
		procs, err := s.db.SubscriptionPurchaseProcessingByReference(ref)
		if err != nil {
			log.Error().Err(err).Msg("could not acquire item subscriptions processing information")
			log.Info().Msg("retrying in 5 minutes...")
			retry = true
			time.Sleep(5 * time.Minute)
			continue
		}
		//only allow one processing entry
		if len(*procs) != 1 {
			if len(*procs) > 1 {
				log.Error().Err(err).Msg("more than one processing entries for the same reference date")
				break
			}
			log.Error().Err(err).Msg("there are no processing entries this reference date")
			break
		}
		proc := (*procs)[0]
		if proc.Status != chassis.Pending {
			log.Error().Err(err).Msg("processing is not pending")
			break
		}
		if today.Day() == ProcessingDay {
			//CHANGING STATUS TO PROCESSING
			startTime := time.Now()
			proc.Status = chassis.Processing
			proc.StartedAt = &startTime
			proc.IsProcessingDay = true
			if err = s.db.UpdateSubscriptionPurchaseProcessing(&proc); err != nil {
				log.Error().Err(err).Msg("error updating processing status")
				break
			}

			//CREATING SUBSCRIPTION PURCHASES
			if err = s.db.CreateSubscriptionPurchases(today.Format(DateLayout)); err != nil {
				log.Error().Err(err).Msg("error while creating subscription purchases")

				endTime := time.Now()
				proc.Status = chassis.Error
				proc.EndedAt = &endTime
				if err = s.db.UpdateSubscriptionPurchaseProcessing(&proc); err != nil {
					log.Error().Err(err).Msg("error updating processing status")
				}
				break
			}

			//PROCESS EACH SUBSCRIPTION PURCHASE
			s.processSubscriptionsPurchases(ref)

			//CHANGING STATUS TO COMPLETED
			endTime := time.Now()
			proc.Status = chassis.Completed
			proc.EndedAt = &endTime
			if err = s.db.UpdateSubscriptionPurchaseProcessing(&proc); err != nil {
				log.Error().Err(err).Msg("error updating processing status")
			}
		} else {
			//CHANGING STATUS TO COMPLETE ON DAYS THAT SHOULDN'T HAVE PROCESSING OF SUBSCRIPTIONS
			endTime := time.Now()
			proc.Status = chassis.Completed
			proc.StartedAt = &today
			proc.EndedAt = &endTime
			proc.IsProcessingDay = false
			if err = s.db.UpdateSubscriptionPurchaseProcessing(&proc); err != nil {
				log.Error().Err(err).Msg("error updating processing status")
				break
			}
		}
	}

}

func (s *Server) processSubscriptionsPurchases(ref string) {
	status := "pending"

	subs, err := s.db.SubscriptionPurchasesByReferenceAndStatus(ref, &status)
	if err != nil {
		log.Error().Err(err).Msg("while getting subscription purchases")
		return
	}

	for _, sub := range *subs {
		if err = s.createSubscriptionPurchase(sub); err != nil {
			errStr := err.Error()
			now := time.Now()
			sub.Status = chassis.Error
			sub.ProcessedAt = &now
			sub.Errors = &errStr
			if err = s.db.UpdateSubscriptionPurchase(&sub); err != nil {
				log.Error().Err(err).Msg("while updating subscription purchase with error")
			}
		}
	}
}

func (s *Server) createSubscriptionPurchase(sub model.SubscriptionPurchase) error {
	//STEP 1 - Getting user address
	addr, err := s.userSvc.GetAddress(sub.BuyerID, sub.AddressID)
	if err != nil {
		return err
	}

	//STEP 2 - Getting all items that are due this month.
	subItems, err := s.db.SubscriptionItemsByOwnerAndReference(sub.BuyerID, sub.Reference)
	if err != nil {
		return err
	}

	//STEP 3 - look for any issue with the items (if it cannot be delivered to this address or out of stock)
	subErrors, err := s.CheckInvalidItems(addr, *subItems)
	if err != nil {
		return err
	}
	//STEP 4 - remove any item that an issue occurred
	for _, v := range *subErrors {
		if len(v) > 0 {

		}
	}

	//STEP 6 - Calculate delivery fees for the remaining items
	fees, err := s.CalculateDeliveryFees(*subItems, addr)
	if err != nil {
		return err
	}
	//STEP 7 - group delivery fees by seller
	deliveriesBySeller := make(map[string]model.DeliveryFee)
	deliveries := model.DeliveryFees{}

	for _, fee := range *fees {
		deliveries = append(deliveries, fee)
		deliveriesBySeller[fee.Seller] = fee
	}

	var purchaseItems []types.PurchaseItem
	itemsByOwner := make(map[string][]types.PurchaseItem) //group items by owner

	//STEP 8 - build purchase items
	// TODO: REFACTOR THIS METHOD TO AVOID DUPLICATED CODE
	for _, subItem := range *subItems {
		//checking if item exists
		var itemInfo *it.ItemFullWithLink
		if subItem.ItemType == it.ProductOfferingItem {
			if itemInfo, err = s.itemSvc.ItemFullWithLink(subItem.ItemID, "is-offering-for"); err != nil {
				log.Error().Err(err).Msg("sub item not found: " + subItem.ItemID)
				return err
			}
		} else {
			if itemInfo, err = s.itemSvc.ItemFullWithLink(subItem.ItemID, ""); err != nil {
				log.Error().Err(err).Msg("cart item not found: " + subItem.ItemID)
				return err
			}
		}
		//creating item entry and adding to the collection
		purInfo := fillPurchaseItemFromSubscription(*itemInfo, subItem)

		purchaseItems = append(purchaseItems, purInfo)

		//if the purchase item is part of an order, group it by owner
		//this may consume more memory, but will be faster to create orders and ease the marshaling process
		if purInfo.ProductType == it.ProductOfferingItem.String() || purInfo.ProductType == it.DishItem.String() {
			itemsByOwner[itemInfo.Owner.ID] = append(itemsByOwner[itemInfo.Owner.ID], purInfo)
		}
	}

	//STEP 9 - creating purchases
	purchase := model.Purchase{}
	purchase.BuyerID = sub.BuyerID
	err = purchase.Status.FromString("pending")
	purchase.Items = purchaseItems
	purchase.DeliveryFees = &deliveries
	purchase.Site = &defaultSite

	//creating orders and bookings
	orders, bookings := fillOrdersAndBookings(purchase.BuyerID, itemsByOwner, deliveriesBySeller, purchaseItems, addr)

	if err = s.db.CreatePurchase(&purchase, &orders, &bookings); err != nil {
		return err
	}

	// STEP 10 - triggering events
	chassis.Emit(s, events.PurchaseCreated, purchase)
	for _, o := range orders {
		chassis.Emit(s, events.OrderCreated, o)
	}
	for _, b := range bookings {
		chassis.Emit(s, events.BookingCreated, b)
	}

	// STEP 10 - updating stock availability
	for _, item := range *subItems {
		if err = s.itemSvc.UpdateItemAvailability(item.ItemID, -item.Quantity); err != nil {
			//TODO: do nothing or log
		}
	}

	//STEP 11 - updating subscription purchase status
	now := time.Now()
	sub.PurchaseID = &purchase.Id
	sub.Status = chassis.Completed
	sub.ProcessedAt = &now
	if err = s.db.UpdateSubscriptionPurchase(&sub); err != nil {
		return err
	}

	//STEP 12 - triggering payments
	intents, _ := s.paymentSvc.CreatePaymentIntent(purchase)
	if intents != nil {
		paid := true
		//checking if all payments are successful
		for _, intent := range *intents {
			if intent.Status != "success" {
				paid = false
			}
		}

		//if true, set purchase status to 'completed'
		if paid == true {
			if err = purchase.Status.FromString("completed"); err != nil {
				return err
			}
			if err = s.db.UpdatePurchase(&purchase); err != nil {
				return err
			}
		}
	}

	//STEP 13 - triggering notification emails
	purchaseView := model.FullView(&purchase, &orders, &bookings, intents)
	//this routine will trigger email notifications separately
	go s.sendPurchaseCreatedNotifications(purchaseView)

	return nil
}

func (s *Server) CalculateDeliveryFees(items []model.SubscriptionItem, addr *usr.Address) (*[]model.DeliveryFee, error) {
	sellersIds := map[string][]it.ItemFullWithLink{}
	subItems := map[string]model.SubscriptionItem{}
	//get all item ids of products
	for _, i := range items {
		if i.ItemType == it.ProductOfferingItem || i.ItemType == it.DishItem {
			subItems[i.ItemID] = i
		}
	}

	uniqueItemIds := []string{}
	for k, _ := range subItems {
		uniqueItemIds = append(uniqueItemIds, k)
	}
	//TODO: unmarshall of FullItem is not complete
	itemsInfo, err := s.itemSvc.GetItems(uniqueItemIds, "")
	if err != nil {
		return nil, err
	}

	for _, i := range *itemsInfo {
		sellersIds[i.Owner.ID] = append(sellersIds[i.Owner.ID], i)
	}
	uniqueSellersIds := []string{}
	for k, _ := range sellersIds {
		uniqueSellersIds = append(uniqueSellersIds, k)
	}

	fees, err := s.userSvc.GetDeliveryFees(uniqueSellersIds)
	if err != nil {
		return nil, err
	}

	deliveryFees := []model.DeliveryFee{}

	for k, v := range *fees {
		sellerItems := sellersIds[k]
		var sum int
		for _, i := range sellerItems {
			price := i.Attrs["price"].(float64)
			sum += int(price) * subItems[i.ID].Quantity //* quantity
		}
		if sum > v.FreeDeliveryAbove {
			continue
		}
		deliveryFees = append(deliveryFees, model.DeliveryFee{
			Seller:   v.Owner,
			Price:    v.NormalOrderPrice,
			Currency: v.Currency,
		})
	}
	return &deliveryFees, nil
}

func (s *Server) CheckInvalidItems(addr *usr.Address, items []model.SubscriptionItem) (*map[string][]string, error) {
	ids := []string{}
	subMap := map[string]model.SubscriptionItem{}
	subErrors := map[string][]string{}

	for _, i := range items {
		ids = append(ids, i.ItemID)
		subMap[i.ItemID] = i
	}

	itemInfo, err := s.itemSvc.GetItems(ids, "")
	if err != nil {
		return nil, err
	}

	for _, info := range *itemInfo {
		//check if product can be delivered for the user default address
		//we won't check if this is an anonymous cart (userID == "", therefore addr == nil)
		if info.ItemType == it.ProductOfferingItem && addr != nil {
			rawZones, ok := info.Attrs["availability_zones"]
			if ok {
				//zones available
				ids := []int{}
				zones := rawZones.([]interface{})
				for _, zone := range zones {
					z := zone.(map[string]interface{})
					ids = append(ids, int(z["reference"].(float64)))
				}
				//check if user address is inside of at least one of available delivery zones
				ok, err := s.searchSvc.CheckRegions(addr.Coordinates.Latitude, addr.Coordinates.Longitude, ids)
				if err != nil {
					return nil, err
				}
				//case negative, append a cart error to be shown on pre-checkout screen
				if !*ok {
					subItemID := subMap[info.ID].ItemID
					subErrors[subItemID] = append(subErrors[subItemID],
						fmt.Sprintf("item '%s' cannot be delivered to your location", info.ID))
				}

			}
		}
		if err = cartUtils.IsAvailableForSale(subMap[info.ID].Quantity, info); err != nil {
			cartItemID := subMap[info.ID].ID
			subErrors[cartItemID] = append(subErrors[cartItemID],
				fmt.Sprintf("item '%s' is not available for sale: %v", info.ID, err))
		}
	}
	return &subErrors, nil
}
