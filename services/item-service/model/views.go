package model

import (
	"time"

	"github.com/lib/pq"
	"github.com/veganbase/backend/services/item-service/model/types"
	user_model "github.com/veganbase/backend/services/user-service/model"
)

// ItemSummary is a view of an item that is used for listing multiple
// items for summary pages.
type ItemSummary struct {
	ID              string         `json:"id" db:"id"`
	ItemType        ItemType       `json:"item_type" db:"item_type"`
	Slug            string         `json:"slug" db:"slug"`
	Lang            string         `json:"lang" db:"lang"`
	Name            string         `json:"name" db:"name"`
	Description     string         `json:"description" db:"description"`
	FeaturedPicture string         `json:"featured_picture" db:"featured_picture"`
	Pictures        pq.StringArray `json:"pictures" db:"pictures"`
	Tags            pq.StringArray `json:"tags" db:"tags"`
	URLs            types.URLMap   `json:"urls" db:"urls"`
}

// Summary generates a summary view of an item from a database model,
// but the more common use case is to create ItemSummary values by
// scanning directly from database query results.
func Summary(item *Item) *ItemSummary {
	return &ItemSummary{
		ID:              item.ID,
		ItemType:        item.ItemType,
		Slug:            item.Slug,
		Lang:            item.Lang,
		Name:            item.Name,
		Description:     item.Description,
		FeaturedPicture: item.FeaturedPicture,
		Pictures:        item.Pictures,
		Tags:            item.Tags,
		URLs:            item.URLs,
	}
}

// ItemFullFixed is a view of the fixed fields of a full view of an
// item.
type ItemFullFixed struct {
	ID              string                `json:"id"`
	ItemType        ItemType              `json:"item_type"`
	Slug            string                `json:"slug"`
	Lang            string                `json:"lang"`
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	FeaturedPicture string                `json:"featured_picture"`
	Pictures        []string              `json:"pictures"`
	Tags            []string              `json:"tags"`
	URLs            types.URLMap          `json:"urls"`
	Creator         *user_model.Info      `json:"creator"`
	Approval        types.ApprovalState   `json:"approval"`
	Owner           *user_model.Info      `json:"owner"`
	Ownership       types.OwnershipStatus `json:"ownership"`
}

// ItemFull is a full view of an item.
type ItemFull struct {
	ItemFullFixed
	Attrs       types.AttrMap `json:"-"`
	Links       []interface{} `json:"-"`
	Upvotes     int           `json:"upvotes"`
	UserUpvoted bool          `json:"user_upvoted"`
	Rank        float64       `json:"rank"`
	Collections []string      `json:"collections"`
}

type ItemFullWithLink struct {
	ItemFullFixed
	Attrs types.AttrMap    `json:"-"`
	Link  *LinkLinkWithFull `json:"link,omitempty"`
}

// FullView generates a full view of an item.
func FullView(item *Item, creator, owner *user_model.Info, upvotes int, upvoted bool, rank float64, collections []string) *ItemFull {
	view := ItemFull{}
	view.ID = item.ID
	view.ItemType = item.ItemType
	view.Slug = item.Slug
	view.Lang = item.Lang
	view.Name = item.Name
	view.Description = item.Description
	view.FeaturedPicture = item.FeaturedPicture
	view.Pictures = item.Pictures
	view.Tags = item.Tags
	view.URLs = item.URLs
	view.Creator = creator
	view.Approval = item.Approval
	view.Owner = owner
	view.Ownership = item.Ownership
	view.Attrs = item.Attrs
	view.Links = nil
	view.Upvotes = upvotes
	view.UserUpvoted = upvoted
	view.Rank = rank
	view.Collections = collections
	return &view
}

// ItemFixedValidate is a view of an item containing all the fixed
// fields that can be modified and require validation.
type ItemFixedValidate struct {
	ItemType        ItemType     `json:"item_type"`
	Lang            string       `json:"lang"`
	Name            string       `json:"name"`
	Description     string       `json:"description"`
	FeaturedPicture string       `json:"featured_picture"`
	Pictures        []string     `json:"pictures"`
	Tags            []string     `json:"tags"`
	URLs            types.URLMap `json:"urls"`
}

// FixedValidate generates a view of an item from a database model
// that contains all the fixed fields that can be modified and require
// validation.
func FixedValidate(item *Item) *ItemFixedValidate {
	return &ItemFixedValidate{
		ItemType:        item.ItemType,
		Lang:            item.Lang,
		Name:            item.Name,
		Description:     item.Description,
		FeaturedPicture: item.FeaturedPicture,
		Pictures:        item.Pictures,
		Tags:            item.Tags,
		URLs:            item.URLs,
	}
}

// LinkLinkWithSummary is a view of an inter-item link including summary
// information for the target item.
type LinkLinkWithSummary struct {
	ID        string       `json:"id"`
	InverseID string       `json:"inverse_id"`
	Origin    string       `json:"origin"`
	Target    *ItemSummary `json:"target"`
	LinkType  string       `json:"link_type"`
	Owner     string       `json:"owner"`
	CreatedAt time.Time    `json:"created_at"`
}

// LinkWithSummary returns a view of an inter-item link that includes
// summary information for the link's target item.
func LinkWithSummary(link *Link, target *ItemSummary) *LinkLinkWithSummary {
	return &LinkLinkWithSummary{
		ID:        link.ID,
		InverseID: link.InverseID,
		Origin:    link.Origin,
		Target:    target,
		LinkType:  link.LinkType,
		Owner:     link.Owner,
		CreatedAt: link.CreatedAt,
	}
}

// LinkLinkWithFull is a view of an inter-item link including full
// information for the target item.
type LinkLinkWithFull struct {
	ID        string    `json:"id"`
	InverseID string    `json:"inverse_id"`
	Origin    string    `json:"origin"`
	Target    *ItemFull `json:"target"`
	LinkType  string    `json:"link_type"`
	Owner     string    `json:"owner"`
	CreatedAt time.Time `json:"created_at"`
}

// LinkWithFull returns a view of an inter-item link that includes
// full information for the link's target item.
func LinkWithFull(link *Link, target *ItemFull) *LinkLinkWithFull {
	return &LinkLinkWithFull{
		ID:        link.ID,
		InverseID: link.InverseID,
		Origin:    link.Origin,
		Target:    target,
		LinkType:  link.LinkType,
		Owner:     link.Owner,
		CreatedAt: link.CreatedAt,
	}
}
