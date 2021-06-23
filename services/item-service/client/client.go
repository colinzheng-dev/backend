package client

import (
	"github.com/veganbase/backend/services/item-service/model"
	"github.com/veganbase/backend/services/item-service/model/types"
)

// Location is a latitude, longitude pair.
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// SearchInfo represents the information that the search service needs
// about an item to build its indexes.
type SearchInfo struct {
	ItemType    model.ItemType      `json:"item_type"`
	Approval    types.ApprovalState `json:"approval"`
	Location    *Location           `json:"location,omitempty"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Content     string              `json:"content,omitempty"`
	Tags        []string            `json:"tags"`
}


// Client is the service client API for the item service.
type Client interface {
	IDs() ([]string, error)
	SearchInfo(id string) (*SearchInfo, error)
	ItemInfo(id string) (*model.Item, error)
	ItemFullWithLink(id, linkType string) (*model.ItemFullWithLink, error)
	UpdateItemAvailability(itemId string, quantity int)  error
	GetItems(ids []string, linkType string) (*[]model.ItemFullWithLink, error)
	GetItemsInfo(ids []string) (map[string]*model.Info, error)
}
