package db

import (
	"errors"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/item-service/model"
	"github.com/veganbase/backend/services/item-service/model/types"
)

// ErrItemNotFound is the error returned when an attempt is made to
// access or manipulate a item with an unknown ID.
var ErrItemNotFound = errors.New("item ID not found")

// ErrReadOnlyField is the error returned when an attempt is made to
// update a read-only field for a item (e.g. slug, creation time).
var ErrReadOnlyField = errors.New("attempt to modify read-only field")

// ErrItemNotOwned is the error returned when an attempt is made to
// access or manipulate a item that the user does not own.
var ErrItemNotOwned = errors.New("user does not own the item")

// ErrClaimNotFound is the error returned when an attempt is made to
// access or manipulate an item ownership claim with an unknown ID.
var ErrClaimNotFound = errors.New("ownership claim ID not found")

// ErrClaimNotOwned is the error returned when an attempt is made to
// access or manipulate an item ownership claim that the user does not
// own.
var ErrClaimNotOwned = errors.New("user does not own the ownership claim")

// ErrLinkNotFound is the error returned when an attempt is made to
// access or manipulate an inter-item link with an unknown ID.
var ErrLinkNotFound = errors.New("inter-item link ID not found")

// ErrUnknownLinkType is the error returned when an attempt is made to
// create an inter-item link of an unknown link type.
var ErrUnknownLinkType = errors.New("unknown link type")

// ErrInverseLinkType is the error returned when an attempt is made to
// create an inter-item link of an inverse link type directly (inverse
// links are created automatically when a forward link is created, and
// cannot be created directly).
var ErrInverseLinkType = errors.New("cannot directly create link of inverse link type")

// ErrLinkTypeRequiresUniqueOrigin is the error returned when an
// attempt is made to create an inter-item link for a link type with a
// unique origin requirement where a link for the requested origin
// already exists.
var ErrLinkTypeRequiresUniqueOrigin = errors.New("link type requires unique origin and incompatible link already exists")

// ErrLinkOwnershipInvalid is the error returned when an attempt is
// made to create an inter-item link between items where the creating
// user doesn't own the origin or target items (or both), depending on
// what the link ownership class requires.
var ErrLinkOwnershipInvalid = errors.New("item ownership doesn't match link type requirements")

// ErrLinkTargetNotFound is the error returned when an attempt is made
// to create an inter-item link to an unknown target item.
var ErrLinkTargetNotFound = errors.New("link target item ID not found")

// ErrBadLinkOriginType is the error returned when an attempt is made
// to create an inter-item link with an origin whose item type does
// not match the requirements of the link type.
var ErrBadLinkOriginType = errors.New("link origin type not allowed for link type")

// ErrBadLinkTargetType is the error returned when an attempt is made
// to create an inter-item link with a target whose item type does not
// match the requirements of the link type.
var ErrBadLinkTargetType = errors.New("link target type not allowed for link type")

// ErrFailedToCreateLink is the error returned when some unidentified
// problem causes the creation of an inter-item link to fail.
var ErrFailedToCreateLink = errors.New("failed to create inter-item link")

// ErrItemCollectionNotFound is the error returned when an attempt is
// made to access or manipulate an item collection with an unknown ID.
var ErrItemCollectionNotFound = errors.New("item collection not found")

// ErrCollectionNameAlreadyExists is the error returned when an
// attempt is made to create an iotem collection with a non-unique
// name.
var ErrCollectionNameAlreadyExists = errors.New("collection name is already in use")

// ErrItemCollectionNotOwned is the error returned when an attempt is
// made to access or manipulate a item collection that the user does
// not own.
var ErrItemCollectionNotOwned = errors.New("user does not own the item collection")

var ErrStatisticNotFound = errors.New("statistic not found")

// SearchParams represents the search parameters that are accessible
// directly from the items database. (Full-text and geo search are
// handled separately.)
type SearchParams struct {
	ItemTypes *[]model.ItemType
	Approval  *[]types.ApprovalState
	Owner     []string
	Tag       *string
	Ids       *[]string
	SortBy    *chassis.Sorting
}

// DB describes the database operations used by the item service.
type DB interface {
	// Retrieval functions for individual items: by ID, by slug and by
	// ID or slug, whichever matches.
	ItemByID(id string) (*model.Item, error)
	ItemBySlug(slug string) (*model.Item, error)
	ItemByIDOrSlug(idOrSlug string) (*model.ItemWithStatistics, error)

	// SummaryItems and FullItems gets lists of items in reverse
	// creation date order, either as item summaries or full items,
	// optionally filtered by an item type, tag and/or approval status
	// and paginated.
	SummaryItems(params *SearchParams, filterIDs []string,
		collIDs []string, pagination *chassis.Pagination) ([]*model.ItemSummary, *uint, error)
	FullItems(params *SearchParams, filterIDs []string,
		collIDs []string, pagination *chassis.Pagination) ([]*model.ItemWithStatistics, *uint, error)

	// ItemNames gets the names of a list of items identified by their
	// IDs.
	ItemNames(ids []string) (map[string]string, error)

	// ItemIDs gets all the existing item IDs.
	ItemIDs() ([]string, error)

	// CreateItem creates a new item.
	CreateItem(item *model.Item) error

	// UpdateItem updates the item's details in the database. The id,
	// item_type, slug, approval, creator, owner, ownership and
	// created_at fields are read-only using this method.
	UpdateItem(item *model.Item, allowedOwner []string) error

	// UpdateAvailability updates the item's availability. More precisely, the attrs column.
	// The other fields won't be affected because the Patch method is not used. item.UpdateAvailability
	// is used instead.
	UpdateAvailability(item *model.Item) error

	// DeleteItem performs the actual deletion of a item. If the owner
	// user ID parameter is non-empty, then the item will only be
	// deleted if it is owned by the given user ID. Returns a list of
	// picture links used by the item.
	// TODO: ADD SOME SORT OF ARCHIVAL MECHANISM INSTEAD.
	Info(ids []string) (map[string]model.Info, error)
	DeleteItem(id string, allowedOwner []string) ([]string, error)

	// TagsForUser returns all the tags used in items owned by a user.
	TagsForUser(userID string) ([]string, error)

	// UpdateItemApproval updates an item's approval state in the
	// database.
	UpdateItemApproval(id string, approval types.ApprovalState) error

	// UpdateItemOwnership updates an item's ownership state in the
	// database.
	UpdateItemOwnership(id string, owner string, ownership types.OwnershipStatus) error

	// CreateClaim creates a new ownership claim.
	CreateClaim(claim *model.Claim) error

	// ClaimByID looks up an ownership claim.
	ClaimByID(id string) (*model.Claim, error)

	// Claims lists outstanding ownership claims, optionally filtering
	// by user or claim status, with pagination.
	Claims(allowedOwner []string, approval *types.ApprovalState,
		page, perPage uint) ([]model.Claim, *uint, error)

	// DeleteClaim deletes an ownership claim.
	DeleteClaim(id string, allowedOwners []string) error

	// UpdateClaim updates the status of an ownership claim.
	UpdateClaim(claim *model.Claim) error

	// Look up an inter-item link type by name.
	LinkTypeByName(name string) (*model.LinkType, error)

	// Look up all link types that may originate from a given item type.
	LinkTypesByOrigin(origin model.ItemType) ([]model.LinkType, error)

	// LinkByID looks up an inter-item link by its ID.
	LinkByID(id string) (*model.Link, error)

	// LinksByOriginID looks up inter-item links originating from a
	// given item by the item ID.
	LinksByOriginID(id string, page, perPage uint) (*[]model.Link, *uint, error)

	// CreateLink creates an inter-item link.
	CreateLink(linkType string, originID string, targetID string,
		userID string, allowedOwners []string) (*model.Link, error)

	// DeleteLink deletes an inter-item link.
	DeleteLink(linkID string, userID string, allowedOwners []string) error

	// CreateCollection creates a new item collection.
	CreateCollection(req *model.ItemCollectionInfo) error

	// CollectionViewByName retrieves all the information for a single
	// item collection.
	CollectionViewByName(id string) (*model.ItemCollectionView, error)

	// CollectionViews lists item collections.
	CollectionViews(owners []string, page, perPage uint) ([]*model.ItemCollectionView, *uint, error)

	// DeleteItemCollections deleted an item collection. If the owner
	// user ID parameter is non-empty, then the item will only be
	// deleted if it is owned by the given user ID.
	// TODO: ADD SOME SORT OF ARCHIVAL MECHANISM INSTEAD.
	DeleteItemCollection(collName string, allowedOwner []string) error

	// AddItemToCollection adds an item to a manual item collection.
	AddItemToCollection(collName string, itemID string,
		before *string, after *string, allowedOwner []string) error

	// DeleteItemFromCollection deletes an item from a manual item
	// collection.
	DeleteItemFromCollection(collName string, itemID string, allowedOwner []string) error

	// SaveEvent saves an event to the database.
	SaveEvent(topic string, eventData interface{}, inTx func() error) error

	FullItemsByIDs(IDs []string) ([]*model.Item, error)

	CollectionViewsByOwners(owners []string) ([]*model.ItemCollectionView, error)

	GetItemTypeInfo() (*[]model.ItemTypeInfo, error)

	CollectionsNamesByItemId(ids []string) (*map[string][]string, error)

	SetItemRank(itemId string, rank float64) error
	SetItemUpvotes(itemId string, upvotes int) error
}

//go:generate go-bindata -pkg db -o migrations.go migrations/...
