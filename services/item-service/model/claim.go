package model

import (
	"time"

	"github.com/veganbase/backend/services/item-service/model/types"
)

// Claim represents a claim on ownership of an item by a user.
type Claim struct {
	// Claim claim ID.
	ID string `db:"id"`

	// ID of the user or organisation making the claim.
	OwnerID string `db:"owner_id"`

	// ID of the item being claimed.
	ItemID string `db:"item_id"`

	// Claim approval state.
	Status types.ApprovalState `db:"status"`

	// Creation time of claim.
	CreatedAt time.Time `db:"created_at"`
}

// ClaimView is a view of a claim incorporating optional extra
// information about the user and item.
type ClaimView struct {
	ID         string              `json:"id"`
	OwnerID    string              `json:"owner_id,omitempty"`
	OwnerName  string              `json:"owner_name,omitempty"`
	OwnerEmail string              `json:"owner_email,omitempty"`
	ItemID     string              `json:"item_id"`
	ItemName   string              `json:"item_name"`
	Status     types.ApprovalState `json:"status"`
	CreatedAt  time.Time           `json:"created_at"`
}

// ViewClaim creates a view of a claim from the claim data, ready for
// the additional fields to be filled in.
func ViewClaim(claim *Claim, includeOwner bool) *ClaimView {
	view := ClaimView{
		ID:        claim.ID,
		ItemID:    claim.ItemID,
		Status:    claim.Status,
		CreatedAt: claim.CreatedAt,
	}
	if includeOwner {
		view.OwnerID = claim.OwnerID
	}
	return &view
}
