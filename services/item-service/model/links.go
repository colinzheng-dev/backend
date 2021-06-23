package model

import (
	"time"

	"github.com/lib/pq"
	"github.com/veganbase/backend/services/item-service/model/types"
)

// LinkType represents an inter-item link type.
type LinkType struct {
	// Name of the link type (used as foreign key in link definitions).
	Name string `db:"name"`

	// OriginType is a list of permissible origin types for links of
	// this type.
	OriginType pq.StringArray `db:"origin_type"`

	// TargetType is a list of permissible target types for links of
	// this type.
	TargetType pq.StringArray `db:"target_type"`

	// Does this link type require a unique origin item?
	UniqueOrigin bool `db:"origin_unique"`

	// The ownership class of the link type.
	Ownership types.LinkOwnershipClass `db:"ownership"`

	// Is this link type an inverse? (As opposed to a forward link
	// type).
	IsInverse bool `db:"is_inverse"`

	// The inverse (if any) of this link type.
	Inverse string `db:"inverse"`
}

// Link represents a single inter-item link.
type Link struct {
	// Primary key ID for link (of the form "lnk_<random>").
	ID string `db:"id" json:"id"`

	// ID for inverse link to this one.
	InverseID string `db:"inverse_id" json:"inverse_id"`

	// Item ID of origin item.
	Origin string `db:"origin" json:"origin"`

	// Item ID of target item.
	Target string `db:"target" json:"target"`

	// Type of link.
	LinkType string `db:"link_type" json:"link_type"`

	// Link owner.
	Owner string `db:"owner" json:"owner"`

	// Link creation timestamp.
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// AllowsOrigin determines whether a particular origin item type is
// allowed for a link type.
func (linkType *LinkType) AllowsOrigin(origin ItemType) bool {
	if len(linkType.OriginType) == 0 {
		return true
	}
	cmp := origin.String()
	for _, t := range linkType.OriginType {
		if t == cmp {
			return true
		}
	}
	return false
}

// AllowsTarget determines whether a particular target item type is
// allowed for a link type.
func (linkType *LinkType) AllowsTarget(target ItemType) bool {
	if len(linkType.TargetType) == 0 {
		return true
	}
	cmp := target.String()
	for _, t := range linkType.TargetType {
		if t == cmp {
			return true
		}
	}
	return false
}
