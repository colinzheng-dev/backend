package db

import (
	"errors"
	itemModel "github.com/veganbase/backend/services/item-service/model"
	itemTypes "github.com/veganbase/backend/services/item-service/model/types"
	"github.com/veganbase/backend/services/search-service/model"
)
var ErrCountryNotFound = errors.New("country not found")
var ErrStateNotFound = errors.New("state not found")

// DB describes the database operations used by the search service.
type DB interface {
	// Geo performs a geo-search within a given distance of a latitude,
	// longitude longitude point, returning a list of item IDs in order
	// of distance.
	Geo(lat, lon float64, dist float64,
		itemType *itemModel.ItemType,
		approval *[]itemTypes.ApprovalState) ([]string, error)

	// FullText performs a full text search for a given query string,
	// returning a list of item IDs in order of relevance.
	FullText(query string,
		itemType *itemModel.ItemType,
		approval *[]itemTypes.ApprovalState) ([]string, error)

	// Region performs a named region search, returning a list of item
	// IDs lying within the boundaries of the region.
	Region(name string) ([]string, error)

	// AddGeo adds geolocation information to the search index for an
	// item.
	AddGeo(id string,
		itemType itemModel.ItemType,
		approval itemTypes.ApprovalState,
		latitude, longitude float64) error

	// AddFullText adds full-text information to the search index for an
	// item.
	AddFullText(id string,
		itemType itemModel.ItemType,
		approval itemTypes.ApprovalState,
		name, description, content string, tags []string) error

	// ItemRemoved deletes search index information for an item.
	ItemRemoved(id string)

	//GetItemsInsideRegion return all items that are located inside an specific region
	GetItemsInsideRegion(regionType, regionReference string,
		itemType *itemModel.ItemType,
		approval *[]itemTypes.ApprovalState ) ([]string, error)

	IsInsideRegions(latitude, longitude float64, references []int ) (*bool, error)

	Countries() (*[]model.Region, error)
	CountryByID(countryID string) (*model.Region, error)
	States(countryID string ) (*[]model.Region, error)
	StateByID(stateID string) (*model.Region, error)

	// SaveEvent saves an event to the database.
	CreateErrorLog(log *model.ErrorLog) error
	SaveEvent(label string, eventData interface{}, inTx func() error) error
}

//go:generate go-bindata -pkg db -o migrations.go migrations/...
