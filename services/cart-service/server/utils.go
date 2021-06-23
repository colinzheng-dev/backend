package server

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/veganbase/backend/services/cart-service/model"
	it "github.com/veganbase/backend/services/item-service/model"
	usr "github.com/veganbase/backend/services/user-service/model"
)

var dateLayout string = "2006-01-02"

var ErrInvalidItemType = errors.New("cannot be added to cart")
var ErrOutOfStock = errors.New("out of stock")
var ErrNotAvailable = errors.New("not for sale")
var ErrCannotDeliver = errors.New("cannot be delivered to your location")

// isAvailable checks if the item is available and if it has enough available in stock
// TODO: check other types (like media) that can be added to a cart
func IsAvailableForSale(quantity int, itemInfo it.ItemFullWithLink) error {
	switch itemInfo.ItemType {
	default:
		return ErrInvalidItemType
	case it.ProductOfferingItem, it.DishItem:
		//check if desired quantity is less than in stock availability
		inStock := itemInfo.Attrs["available_quantity"]
		if inStock != nil {
			if quantity > int(inStock.(float64)) {
				return ErrOutOfStock
			}
		}

	case it.OfferItem:
		//first check if the offer is available
		available := itemInfo.Attrs["is_available"]
		if available == nil {
			return ErrNotAvailable
		}
		//if positive, check if desired quantity is less than in stock availability
		if available.(bool) {
			inStock := itemInfo.Attrs["available_quantity"]
			if inStock != nil {
				if quantity <= int(inStock.(float64)) {
					return nil
				}
			}
			return ErrOutOfStock
		}
	case it.RoomItem:
		//check if the offer is available
		available := itemInfo.Attrs["is_available"]
		if available != nil {
			if itemInfo.Attrs["is_available"].(bool) {
				return nil
			}
		}
		return ErrNotAvailable
	}
	return nil
}

func (s *Server) CheckInvalidItems(userID string, items []model.CartItem) (*map[int][]string, error) {
	ids := []string{}
	cartMap := map[string]model.CartItem{}
	cartErrors := map[int][]string{}

	for _, i := range items {
		ids = append(ids, i.ItemID)
		cartMap[i.ItemID] = i
	}

	itemInfo, err := s.itemSvc.GetItems(ids, "")
	if err != nil {
		return nil, err
	}

	var addr *usr.Address
	if userID != "" {
		addr, err = s.userSvc.GetDefaultAddress(userID)
		if err != nil && err.Error() != "default address not found" {
			return nil, err
		}
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
					cartItemID := cartMap[info.ID].ID
					cartErrors[cartItemID] = append(cartErrors[cartItemID],
						fmt.Sprintf("item '%s' cannot be delivered to your location", info.ID))
				}

			}
		}
		if err = IsAvailableForSale(cartMap[info.ID].Quantity, info); err != nil {
			cartItemID := cartMap[info.ID].ID
			cartErrors[cartItemID] = append(cartErrors[cartItemID],
				fmt.Sprintf("item '%s' is not available for sale: %v", info.ID, err))
		}
	}
	return &cartErrors, nil
}

//getNumberOfDays calculates the number of days between a start-date and an end-date to be assigned as quantity
func GetNumberOfDays(startDate, endDate string) (int, error) {
	var t1, t2 time.Time
	var err error

	if t1, err = time.Parse(dateLayout, strings.Split(startDate, "T")[0]); err != nil {
		return 0, err
	}
	if t2, err = time.Parse(dateLayout, strings.Split(endDate, "T")[0]); err != nil {
		return 0, err
	}
	return int(t2.Sub(t1).Hours()) / 24, nil
}

func ValidatePeriod(start, end string) error {
	var t1, t2 *time.Time
	var err error

	if t1, err = ValidateDatetime(start); err != nil {
		return errors.New("error validating period: " + err.Error())
	}
	if t2, err = ValidateDatetime(end); err != nil {
		return errors.New("error validating period: " + err.Error())
	}
	if t2.Sub(*t1).Seconds() < 0 {
		return errors.New("start date is greater than end date")
	}
	return nil
}

func ValidateDatetime(date string) (*time.Time, error) {
	t1, err := time.Parse(time.RFC3339, date+"Z")
	if err != nil {
		return nil, err
	}
	return &t1, nil
}

func (s *Server) CheckDeliveryRegion(latitude, longitude float64, itemInfo it.ItemFullWithLink) error {
	rawZones, ok := itemInfo.Attrs["availability_zones"]
	references := []int{}
	if ok {
		zones := rawZones.([]interface{})
		for _, zone := range zones {
			z := zone.(map[string]interface{})
			references = append(references, int(z["reference"].(float64)))
		}
		if len(references) > 0 {
			valid, err := s.searchSvc.CheckRegions(latitude, longitude, references)
			if err != nil {
				return err
			}
			if !*valid {
				return ErrCannotDeliver
			}
		}
	}
	return nil
}

//groups products (product offering and dishes) by seller and calculates their delivery fee
func (s *Server) CalculateDeliveryFees(items []model.CartItem) (*[]model.DeliveryFee, error) {
	sellersIds := map[string][]it.ItemFullWithLink{}
	cartItems := map[string]model.CartItem{}
	//get all item ids of products
	for _, i := range items {
		if i.Type == it.ProductOfferingItem || i.Type == it.DishItem {
			cartItems[i.ItemID] = i
		}
	}

	uniqueItemIds := []string{}
	for k, _ := range cartItems {
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
			sum += int(price) * cartItems[i.ID].Quantity //* quantity
		}
		if sum > v.FreeDeliveryAbove {
			continue
		}
		deliveryFees = append(deliveryFees, model.DeliveryFee{
			Seller:            v.Owner,
			Price:             v.NormalOrderPrice,
			FreeDeliveryAbove: v.FreeDeliveryAbove,
			Currency:          v.Currency,
		})
	}
	return &deliveryFees, nil
}
