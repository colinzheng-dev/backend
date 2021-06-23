package db

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/item-service/model"
	"github.com/veganbase/backend/services/item-service/model/types"
)

// LinkTypeByName looks up an inter-item link type by name.
func (pg *PGClient) LinkTypeByName(name string) (*model.LinkType, error) {
	return linkTypeByName(pg.DB, name)
}

// Transaction wrapper for link type lookup.
func linkTypeByName(q sqlx.Queryer, name string) (*model.LinkType, error) {
	linkType := &model.LinkType{}
	err := sqlx.Get(q, linkType, qLinkTypeByName, name)
	if err == sql.ErrNoRows {
		return nil, ErrUnknownLinkType
	}
	if err != nil {
		return nil, err
	}
	return linkType, nil
}

const qLinkTypeByName = `
SELECT name, origin_type, target_type, origin_unique, ownership,
       is_inverse, inverse
  FROM item_link_types WHERE name = $1`

// LinkTypesByOrigin looks up all inter-item link types that may
// originate from a given item type.
func (pg *PGClient) LinkTypesByOrigin(itemType model.ItemType) ([]model.LinkType, error) {
	return linkTypesByOrigin(pg.DB, itemType)
}

// Transaction wrapper for link type lookup.
func linkTypesByOrigin(q sqlx.Queryer, itemType model.ItemType) ([]model.LinkType, error) {
	linkTypes := []model.LinkType{}
	err := sqlx.Select(q, &linkTypes, qLinkTypesByOrigin, itemType.String())
	if err == sql.ErrNoRows {
		return nil, ErrUnknownLinkType
	}
	if err != nil {
		return nil, err
	}
	return linkTypes, nil
}

const qLinkTypesByOrigin = `
SELECT name, origin_type, target_type, origin_unique, ownership,
       is_inverse, inverse
  FROM item_link_types WHERE origin_type = '{}' OR $1 = ANY(origin_type)`

// LinkByID looks up an inter-item link by its ID.
func (pg *PGClient) LinkByID(id string) (*model.Link, error) {
	return linkByID(pg.DB, id)
}

// Transaction wrapper for inter-item link lookup.
func linkByID(q sqlx.Queryer, id string) (*model.Link, error) {
	link := &model.Link{}
	if err := sqlx.Get(q, link, qLinkBy + "id = $1", id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}
	return link, nil
}

const qLinkBy = `
SELECT id, inverse_id, origin, target, link_type, owner, created_at
  FROM item_links WHERE `

// LinksByOriginID looks up inter-item links originating from a given
// item by the item ID.
func (pg *PGClient) LinksByOriginID(id string, page, perPage uint) (*[]model.Link, *uint, error) {
	links := []model.Link{}
	var total uint

	// Check origin item exists.
	if _, err := pg.ItemByID(id); err == ErrItemNotFound {
		return nil, nil, ErrItemNotFound
	}

	// Retrieve links.
	q := qLinkBy + `origin = '` + id + `' ORDER BY created_at DESC ` + chassis.Paginate(page, perPage)
	if err := pg.DB.Select(&links, q); err != nil {
		return nil, nil, err
	}

	if err := chassis.GetTotalResultsFromQuery(q, &total, pg.DB); err != nil {
		return nil, &total, err
	}
	return &links, &total, nil
}

// CreateLink creates an inter-item link.
func (pg *PGClient) CreateLink(linkTypeName string,
	originID string, targetID string,
	userID string, allowedOwners []string) (*model.Link, error) {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	// Look up link type.
	linkType, err := linkTypeByName(tx, linkTypeName)
	if err != nil {
		return nil, err
	}

	// Check we're not trying to create a link of an inverse link type
	// directly.
	if linkType.IsInverse {
		return nil, ErrInverseLinkType
	}

	// Look up origin and target items.
	origin, err := pg.ItemByID( originID)
	if err == ErrItemNotFound {
		return nil, ErrItemNotFound
	}
	target, err := pg.ItemByID(targetID)
	if err == ErrItemNotFound {
		return nil, ErrLinkTargetNotFound
	}

	// Check origin and target item types compatibility with link type.
	if !linkType.AllowsOrigin(origin.ItemType) {
		return nil, ErrBadLinkOriginType
	}
	if !linkType.AllowsTarget(target.ItemType) {
		return nil, ErrBadLinkTargetType
	}

	// Check unique origin rules.
	if linkType.UniqueOrigin {
		exists, err := linkExists(tx, linkType, originID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrLinkTypeRequiresUniqueOrigin
		}
	}

	// Check link ownership class rules.
	owner, ok := ownershipClassOK(linkType, origin, target, userID, allowedOwners)
	if !ok {
		return nil, ErrLinkOwnershipInvalid
	}

	// Create an ID for the new link and for any inverse link.
	linkID := chassis.NewID("lnk")
	inverseLinkID := ""
	if linkType.Inverse != "" {
		inverseLinkID = chassis.NewID("lnk")
	}
	link := model.Link{
		ID:        linkID,
		InverseID: inverseLinkID,
		Origin:    originID,
		Target:    targetID,
		LinkType:  linkType.Name,
		Owner:     owner,
	}

	// Create the link.
	var rows *sql.Rows
	if linkType.Inverse == "" {
		rows, err = tx.Query(qCreateSingleLink,
			linkID, originID, targetID, linkType.Name, owner)
	} else {
		rows, err = tx.Query(qCreatePairedLinks,
			linkID, inverseLinkID, originID, targetID,
			linkType.Name, linkType.Inverse, owner)
	}
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	err = rows.Scan(&link.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &link, nil
}

const qCreateSingleLink = `
INSERT INTO item_links (id, origin, target, link_type, owner)
VALUES ($1, $2, $3, $4, $5)
RETURNING created_at`

const qCreatePairedLinks = `
INSERT INTO item_links (id, inverse_id, origin, target, link_type, owner)
VALUES
 ($1, $2, $3, $4, $5, $7),
 ($2, $1, $4, $3, $6, $7)
RETURNING created_at`

// DeleteLink deletes an inter-item link.
func (pg *PGClient) DeleteLink(linkID string, userID string, allowedOwners []string) error {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	// Retrieve the link.
	link, err := linkByID(tx, linkID)
	if err == sql.ErrNoRows {
		return ErrLinkNotFound
	}

	// Get link type and origin and target items.
	linkType, err := linkTypeByName(tx, link.LinkType)
	if err != nil {
		return ErrUnknownLinkType
	}
	origin, err := pg.ItemByID(link.Origin)
	if err != nil {
		return ErrItemNotFound
	}
	target, err := pg.ItemByID(link.Target)
	if err != nil {
		return ErrItemNotFound
	}

	// Check that the link type ownership class allows the user to
	// delete the link.
	_, ok := ownershipClassOK(linkType, origin, target, userID, allowedOwners)
	if !ok {
		return ErrLinkOwnershipInvalid
	}

	// Delete the link (automatically deletes inverse link via foreign
	// key "ON CASCADE").
	result, err := tx.Exec(qDeleteLink, linkID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil || rows != 1 {
		return err
	}

	return nil
}

const qDeleteLink = `
DELETE FROM item_links WHERE id = $1`

func ownershipClassOK(linkType *model.LinkType, origin *model.Item,
	target *model.Item, userID string, allowedOwners []string) (string, bool) {
	if len(allowedOwners) == 0 {
		return userID, true
	}

	for _, owner := range allowedOwners {
		ownsOrigin := origin.Owner == owner
		ownsTarget := target.Owner == owner

		switch linkType.Ownership {
		case types.OwnerToOwner:
			if ownsOrigin && ownsTarget {
				return owner, true
			}
		case types.OwnerToAny:
			if ownsOrigin {
				return owner, true
			}
		case types.AnyToOwner:
			if ownsTarget {
				return owner, true
			}
		}
	}
	return "", false
}

func linkExists(q sqlx.Queryer, linkType *model.LinkType, originID string) (bool, error) {
	rows, err := q.Query(qLinkExists, linkType.Name, originID)
	if err != nil {
		return false, err
	}
	if rows.Next() {
		return true, nil
	}
	return false, nil
}

const qLinkExists = `
SELECT id FROM item_links WHERE link_type = $1 AND origin = $2 LIMIT 1`
