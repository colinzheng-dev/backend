package model

import (
	"time"

	"github.com/lib/pq"

	"github.com/veganbase/backend/services/item-service/model/types"
)

//go:generate go run ../tools/gen_item_types/main.go

// Item is the base model for all items.
type Item struct {
	// Unique ID of the item.
	ID string `db:"id"`

	// Item type, e.g. "hotel", "restaurant", "article", "offer", etc.
	ItemType ItemType `db:"item_type"`

	// Slug for use in URLs derived from item name.
	Slug string `db:"slug"`

	// Language used for item text.
	Lang string `db:"lang"`

	// Human-readable name for item.
	Name string `db:"name"`

	// Full description of item.
	Description string `db:"description"`

	// Image displayed as the "featured image" for the item. Must appear
	// in the Pictures array.
	FeaturedPicture string `db:"featured_picture"`

	// Images associated with item.
	Pictures pq.StringArray `db:"pictures"`

	// Item tags: these appear as hashtags in front end.
	Tags pq.StringArray `db:"tags"`

	// URLs associated with item, along with their link types (e.g.
	// website, Facebook, etc.)
	URLs types.URLMap `db:"urls"`

	// Item attributes.
	Attrs types.AttrMap `db:"attrs"`

	// Item approval state.
	Approval types.ApprovalState `db:"approval"`

	// User ID of the user who originally created this item.
	Creator string `db:"creator"`

	// User ID of the registered owner of this item.
	Owner string `db:"owner"`

	// Ownership status of item: this is either "CREATOR" (the initial
	// value where an item is owned by its creator) or "CLAIMED" (when
	// an item has been claimed by a user and the claim has been
	// approved).
	Ownership types.OwnershipStatus `db:"ownership"`

	// Creation time of item.
	CreatedAt time.Time `db:"created_at"`
}

type ItemWithStatistics struct {
	Item
	Rank    float64 `db:"rank"`
	Upvotes int `db:"upvotes"`
}

func (item *Item) checkFeaturedPicture() bool {
	for _, p := range item.Pictures {
		if p == item.FeaturedPicture {
			return true
		}
	}
	return false
}
