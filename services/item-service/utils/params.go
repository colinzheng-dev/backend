package utils

import (
	"net/url"
	"strings"

	"github.com/veganbase/backend/services/item-service/model"
	"github.com/veganbase/backend/services/item-service/model/types"
)

// ItemTypeParam extracts a URL query parameter for an item type.
func ItemTypeParam(qs url.Values, dst **model.ItemType) error {
	s := qs.Get("type")
	if s != "" {
		itemType := model.UnknownItem
		if err := itemType.FromString(s); err != nil {
			return err
		}
		*dst = &itemType
	}
	return nil
}

func ItemsTypeParam(qs url.Values) ([]model.ItemType, error) {
	itemsType := []model.ItemType{}

	s := qs.Get("type")
	ss := strings.Split(s, ",")

	for _, s := range ss {
		if s != "" {
			itemType := model.UnknownItem
			if err := itemType.FromString(s); err != nil {
				return nil, err
			}
			switch itemType {
			case model.VenueItem:
				itemsType = append(itemsType, model.RestaurantItem, model.HotelItem, model.ShopItem, model.CafeItem)
			case model.ExperienceItem:
				itemsType = append(itemsType, model.RoomItem, model.OfferItem, model.ServiceItem)
			case model.ProductItem:
				itemsType = append(itemsType, model.FreshFoodItem,model.PackagedFoodItem, model.CosmeticsItem,
					model.DishItem, model.FashionItem, model.HomewareItem)
			case model.MediaItem:
				itemsType = append(itemsType, model.JobAdItem, model.PostItem, model.RecipeItem, model.ArticleItem)
			default:
				itemsType = append(itemsType, itemType)
			}


		}
	}

	return itemsType, nil
}

// ApprovalParam extracts a URL query parameter for item approval
// statuses.
func ApprovalParam(qs url.Values, dst **[]types.ApprovalState) error {
	s := qs.Get("approval")
	if s == "" {
		approval := []types.ApprovalState{types.Approved}
		*dst = &approval
	} else {
		approvals := []types.ApprovalState{}
		for _, app := range strings.Split(s, ",") {
			approval := types.ApprovalState(types.Pending)
			if err := approval.FromString(app); err != nil {
				return err
			}
			approvals = append(approvals, approval)
		}
		*dst = &approvals
	}
	return nil
}
