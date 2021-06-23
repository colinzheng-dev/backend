package db

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/item-service/model"
)

// CreateCollection creates a new item collection.
func (pg *PGClient) CreateCollection(req *model.ItemCollectionInfo) (err error) {
	// Set up basic item collection entry.
	var collID int
	err = pg.DB.Get(&collID, createColl, strings.TrimSpace(req.Name), req.Owner)
	if err == sql.ErrNoRows {
		return ErrCollectionNameAlreadyExists
	}
	if err != nil {
		return err
	}
	return nil
}

const createColl = `
INSERT INTO item_colls (name, owner)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
RETURNING id`

// Transaction wrapper for item collection lookup.
func collectionByID(q sqlx.Queryer, id string) (*model.ItemCollection, error) {
	coll := &model.ItemCollection{}
	err := sqlx.Get(q, coll, qCollectionByID, id)
	if err != nil {
		return nil, err
	}
	return coll, nil
}

const qCollectionByID = `
SELECT id, name, owner FROM item_colls WHERE id = $1`

// Transaction wrapper for item collection name lookup.
func collectionByName(q sqlx.Queryer, name string) (*model.ItemCollection, error) {
	coll := &model.ItemCollection{}
	err := sqlx.Get(q, coll, qCollectionByName, name)
	if err == sql.ErrNoRows {
		return nil, ErrItemCollectionNotFound
	}
	if err != nil {
		return nil, err
	}
	return coll, nil
}

const qCollectionByName = `
SELECT id, name, owner FROM item_colls WHERE name = $1`

// CollectionViews lists item collections.
func (pg *PGClient) CollectionViews(owners []string,
	page, perPage uint) ([]*model.ItemCollectionView, *uint, error) {
	var total uint
	tx, err := pg.DB.Beginx()
	if err != nil {
		return nil, &total, err
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	names := []string{}
	q := qCollNames + ownerWhere(owners) + oCollNames + chassis.Paginate(page, perPage)
	if err = tx.Select(&names, q); err != nil {
		return nil, &total, err
	}

	views := make([]*model.ItemCollectionView, len(names))
	for i := 0; i < len(names); i++ {
		v, err := collectionViewByName(tx, names[i], false)
		if err != nil {
			return nil, &total, err
		}
		views[i] = v
	}
	if err = chassis.GetTotalResultsFromQuery(q, &total, pg.DB); err != nil {
		return nil, nil, err
	}
	return views, &total, nil
}

// CollectionViews lists item collections.
func (pg *PGClient) CollectionViewsByOwners(owners []string) ([]*model.ItemCollectionView, error) {
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

	query := qCollNames + ownerWhere(owners) + oCollNames

	names := []string{}
	err = tx.Select(&names, query)
	if err != nil {
		return nil, err
	}

	views := make([]*model.ItemCollectionView, len(names))
	for i := 0; i < len(names); i++ {
		v, err := collectionViewByName(tx, names[i], true)
		if err != nil {
			return nil, err
		}
		views[i] = v
	}

	return views, nil
}

const qCollNames = `SELECT name FROM item_colls WHERE `
const oCollNames = ` ORDER BY name `

func ownerWhere(owners []string) string {
	w := `TRUE`
	if owners != nil && len(owners) != 0 {
		w = `owner IN ('` + strings.Join(owners, "', '") + `')`
	}
	return w
}

// CollectionViewByName retrieves all the information for a single
// item collection.
func (pg *PGClient) CollectionViewByName(name string) (*model.ItemCollectionView, error) {
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
	return collectionViewByName(tx, name, true)
}

func collectionViewByName(tx *sqlx.Tx, name string, includeIDs bool) (*model.ItemCollectionView, error) {
	collID, collInfo, err := collectionInfoByName(tx, name)
	if err != nil {
		return nil, err
	}
	ids := []string(nil)
	if includeIDs {
		err = tx.Select(&ids, qManualCollIDs, collID)
		if err != nil {
			return nil, err
		}
	}
	view := model.ItemCollectionView{
		ItemCollectionInfo: *collInfo,
		IDs:                ids,
	}
	return &view, nil
}

const qManualCollIDs = `
SELECT item_id FROM item_colls_items WHERE coll_id = $1 ORDER BY idx`

func collectionInfoByName(q sqlx.Queryer, name string) (int, *model.ItemCollectionInfo, error) {
	coll, err := collectionByName(q, name)
	if err != nil {
		return 0, nil, err
	}

	info := model.ItemCollectionInfo{
		ID:    coll.ID,
		Name:  coll.Name,
		Owner: coll.Owner,
	}
	return coll.ID, &info, nil
}

const qCollectionsByItemId = `
SELECT ici.item_id, string_agg(c.name, ',') as collection_names
FROM item_colls c INNER JOIN item_colls_items ici ON ici.coll_id = c.id `

func (pg *PGClient) CollectionsNamesByItemId(ids []string) (*map[string][]string, error) {
	names := make(map[string][]string)
	var whereClause string

	//initializing map with empty collections
	for _, i := range ids {
		names[i] = []string{}
	}
	if len(ids) > 0 {
		whereClause = "WHERE ici.item_id IN ('" + strings.Join(ids, "','") + "')"
	}
	rows, err := pg.DB.Queryx(qCollectionsByItemId + whereClause + " GROUP BY ici.item_id")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		current := struct {
			Id        string `db:"item_id"`
			CollNames string `db:"collection_names"`
		}{}
		if err = rows.Scan(&current.Id, &current.CollNames); err != nil {
			return nil, err
		}
		collNames := strings.Split(current.CollNames, ",")

		names[current.Id] = collNames
	}
	return &names, nil

}

// DeleteItemCollection deletes an item collection.
// TODO: ADD SOME SORT OF ARCHIVAL MECHANISM INSTEAD.
func (pg *PGClient) DeleteItemCollection(collName string, allowedOwner []string) error {
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

	// Check that item collection exists, so that we can distinguish
	// between bad item collection ID and bad owner below.
	check, err := collectionByName(tx, collName)
	if err != nil {
		return ErrItemCollectionNotFound
	}

	// Check owner.
	if len(allowedOwner) > 0 {
		allowed := false
		for _, o := range allowedOwner {
			if check.Owner == o {
				allowed = true
				break
			}
		}
		if !allowed {
			return ErrItemCollectionNotOwned
		}
	}

	// Try to delete the item collection. (All the associated
	// information for the collection is deleted by Postgres because of
	// cascade constraints.)
	result, err := tx.Exec(qDeleteItemCollection, collName)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrItemCollectionNotOwned
	}
	return nil
}

const qDeleteItemCollection = `
DELETE FROM item_colls WHERE name = $1`

// AddItemToCollection adds an item to a manual item collection.
func (pg *PGClient) AddItemToCollection(collName string, itemID string,
	before *string, after *string, allowedOwner []string) error {
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

	// Check that item collection exists, so that we can distinguish
	// between bad item collection ID and bad owner below.
	coll, err := collectionByName(tx, collName)
	if err != nil {
		return ErrItemCollectionNotFound
	}

	// Check that the item exists.
	_, err = pg.ItemByID(itemID)
	if err != nil {
		return ErrItemNotFound
	}

	// Check owner.
	if len(allowedOwner) > 0 {
		allowed := false
		for _, o := range allowedOwner {
			if coll.Owner == o {
				allowed = true
				break
			}
		}
		if !allowed {
			return ErrItemCollectionNotOwned
		}
	}

	if before != nil && after != nil {
		return errors.New("can't give both 'before' and 'after'")
	}

	var maxIdx *int
	err = tx.Get(&maxIdx, qAddItemMaxIdx, coll.ID)
	if err != nil {
		return err
	}

	newIdx := 0
	if before == nil && after == nil {
		newIdx = 1
		if maxIdx != nil {
			newIdx = *maxIdx + 1
		}
	} else {
		// Find index of marker item.
		marker := before
		if after != nil {
			marker = after
		}
		markerIdx := 0
		err = tx.Get(&markerIdx, qAddItemFindIdx, coll.ID, *marker)
		if err == sql.ErrNoRows {
			err = errors.New("'before' item ID not found")
			if after != nil {
				err = errors.New("'after' item ID not found")
			}
			return err
		}
		if err != nil {
			return err
		}

		if after != nil {
			markerIdx++
		}
		newIdx = markerIdx
		for upd := *maxIdx; upd >= newIdx; upd-- {
			_, err = tx.Exec(qAddItemMakeSpace, coll.ID, upd)
			if err != nil {
				return err
			}
		}
	}

	_, err = tx.Exec(qAddItem, coll.ID, newIdx, itemID)
	return err
}

// No before/after => after last
const qAddItemMaxIdx = `
SELECT MAX(idx) FROM item_colls_items WHERE coll_id = $1`

const qAddItemFindIdx = `
SELECT idx FROM item_colls_items WHERE coll_id = $1 AND item_id = $2`

const qAddItemMakeSpace = `
UPDATE item_colls_items SET idx = idx + 1 WHERE coll_id = $1 AND idx = $2`

const qAddItem = `
INSERT INTO item_colls_items (coll_id, idx, item_id) VALUES ($1, $2, $3)`

// DeleteItemFromCollection deletes an item from a manual item
// collection.
func (pg *PGClient) DeleteItemFromCollection(collName string, itemID string,
	allowedOwner []string) error {
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

	// Check that item collection exists, so that we can distinguish
	// between bad item collection ID and bad owner below.
	coll, err := collectionByName(tx, collName)
	if err != nil {
		return ErrItemCollectionNotFound
	}

	// Check owner.
	if len(allowedOwner) > 0 {
		allowed := false
		for _, o := range allowedOwner {
			if coll.Owner == o {
				allowed = true
				break
			}
		}
		if !allowed {
			return ErrItemCollectionNotOwned
		}
	}

	var maxIdx *int
	err = tx.Get(&maxIdx, qAddItemMaxIdx, coll.ID)
	if err != nil {
		return err
	}

	// Try to delete the item from the item collection.
	var idx int
	err = tx.Get(&idx, deleteItemFromColl, coll.ID, itemID)
	if err != nil {
		return err
	}

	if maxIdx != nil && *maxIdx > 1 {
		for upd := idx + 1; upd <= *maxIdx; upd++ {
			_, err = tx.Exec(fixCollIndexes, coll.ID, upd)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

const deleteItemFromColl = `
DELETE FROM item_colls_items WHERE coll_id = $1 AND item_id = $2
RETURNING idx`

const fixCollIndexes = `
UPDATE item_colls_items SET idx = idx - 1
WHERE coll_id = $1 AND idx = $2`
