package db

import (
	"database/sql"
	"sort"
	"strings"

	"github.com/gosimple/slug"
	"github.com/jmoiron/sqlx"

	// Import Postgres DB driver.

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/item-service/model"
	"github.com/veganbase/backend/services/item-service/model/types"
)

// ItemByID looks up a item by its ID.
func (pg *PGClient) ItemByID(id string) (*model.Item, error) {
	item := &model.Item{}
	if err := pg.DB.Get(item, qItemBy+` id = $1 `, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrItemNotFound
		}
		return nil, err
	}
	return item, nil
}

// ItemBySlug looks up a item by its slug.
func (pg *PGClient) ItemBySlug(slug string) (*model.Item, error) {
	item := &model.Item{}
	if err := pg.DB.Get(item, qItemBy+`slug = $1`, slug); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrItemNotFound
		}
		return nil, err
	}
	return item, nil
}

// ItemByIDOrSlug looks up a item by its ID or slug.
func (pg *PGClient) ItemByIDOrSlug(idOrSlug string) (*model.ItemWithStatistics, error) {
	item := &model.ItemWithStatistics{}
	if err := pg.DB.Get(item, qItemWithStatisticsBy+` i.id = $1 OR i.slug = $1`, idOrSlug); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrItemNotFound
		}
		return nil, err
	}
	return item, nil
}

const qItemBy = `
SELECT i.id, i.item_type, i.slug, i.lang, i.name, i.description,
  i.featured_picture, i.pictures, i.tags, i.urls, i.attrs,
  i.approval, i.creator, i.owner, i.ownership, i.created_at
  FROM items i WHERE `

const qItemSummaryBy = `
SELECT i.id, i.item_type, i.slug, i.lang, i.name, i.description,
  i.featured_picture, i.pictures, i.tags, i.urls
  FROM items i WHERE `

// SummaryItems gets a list of item summaries in reverse creation date
// order, optionally filtered by an item type, tag and/or approval
// status and paginated.
func (pg *PGClient) SummaryItems(params *SearchParams, filterIDs []string,
	collIDs []string, pagination *chassis.Pagination) ([]*model.ItemSummary, *uint, error) {
	results := []*model.ItemSummary{}
	var total uint

	// if at least one list of ids is passed, we'll insert it on the where clause
	var combinedIDs []string
	if params.Ids != nil {
		combinedIDs = intersectIDs(filterIDs, intersectIDs(collIDs, *params.Ids))
	} else {
		combinedIDs = intersectIDs(filterIDs, collIDs)
	}

	var idInClause string
	if len(combinedIDs) > 0 {
		idInClause = ` AND i.id IN ` + chassis.BuildInArgument(combinedIDs)
	}

	q := qItemSummaryBy + paramsWhere(params) + idInClause +
		` ORDER BY i.created_at DESC `

	if pagination != nil {
		q += chassis.Paginate(pagination.Page, pagination.PerPage)
	}

	if err := pg.DB.Select(&results, q); err != nil {
		return nil, &total, err
	}

	// For searches with geo location we keep the order returned by search-service,
	// as long as there is no sort_by argument
	if len(filterIDs) != 0 && params.SortBy == nil {
		order := makeOrderMap(combinedIDs)
		sort.Slice(results, func(i, j int) bool { return order[results[i].ID] < order[results[j].ID] })
	}

	if err := chassis.GetTotalResultsFromQuery(q, &total, pg.DB); err != nil {
		return nil, nil, err
	}

	return results, &total, nil
}

const qItemWithStatisticsBy = `
SELECT i.id, i.item_type, i.slug, i.lang, i.name, i.description,
  i.featured_picture, i.pictures, i.tags, i.urls, i.attrs,
  i.approval, i.creator, i.owner, i.ownership, i.created_at,
  COALESCE(ist.rank,0) as rank, COALESCE(ist.upvotes,0) as upvotes
FROM items i 
LEFT JOIN item_statistics ist ON ist.item_id = i.id
WHERE `

// FullItems gets a list of full items in reverse creation date order,
// optionally filtered by an item type, tag and/or approval status and
// paginated.
func (pg *PGClient) FullItems(params *SearchParams, filterIDs []string,
	collIDs []string, pagination *chassis.Pagination) ([]*model.ItemWithStatistics, *uint, error) {
	results := []*model.ItemWithStatistics{}
	var total uint

	// if at least one list of ids is passed, we'll insert it on the where clause
	//filterIDs are returned by search-service, collIDs are items belonging to a collection and
	//params.Ids are given by the user
	var combinedIDs []string
	if params.Ids != nil {
		combinedIDs = intersectIDs(filterIDs, intersectIDs(collIDs, *params.Ids))
	} else {
		combinedIDs = intersectIDs(filterIDs, collIDs)
	}

	var idInClause string
	if len(combinedIDs) > 0 {
		idInClause = ` AND i.id IN ` + chassis.BuildInArgument(combinedIDs)
	}

	q := qItemWithStatisticsBy + paramsWhere(params) + idInClause +
		paramsOrderBy(params) + " "

	//this is necessary because we call this method internally, and we do not want pagination
	if pagination != nil {
		q += chassis.Paginate(pagination.Page, pagination.PerPage)
	}

	if err := pg.DB.Select(&results, q); err != nil {
		return nil, nil, err
	}

	// For searches with geo location we keep the order returned by search-service,
	// as long as there is no sort_by argument
	if len(filterIDs) != 0 && params.SortBy == nil {
		order := makeOrderMap(combinedIDs)
		sort.Slice(results, func(i, j int) bool { return order[results[i].ID] < order[results[j].ID] })
	}

	if err := chassis.GetTotalResultsFromQuery(q, &total, pg.DB); err != nil {
		return nil, nil, err
	}

	return results, &total, nil
}

func (pg *PGClient) FullItemsByIDs(IDs []string) ([]*model.Item, error) {
	results := []*model.Item{}

	q := qItemBy + ` i.id IN (?)` +
		` ORDER BY i.created_at DESC `
	q, args, err := sqlx.In(q, IDs)
	if err != nil {
		return nil, err
	}

	q = pg.DB.Rebind(q)
	err = pg.DB.Select(&results, q, args...)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func paramsWhere(ps *SearchParams) string {
	es := []string{}

	//if ps.Ids != nil && len(*ps.Ids) > 0 {
	//	es = append(es, `i.id IN ('` +strings.Join(*ps.Ids, "','") + `')`)
	//}

	if ps.ItemTypes != nil && len(*ps.ItemTypes) > 0 {
		itemTypes := []string{}
		for _, item := range *ps.ItemTypes {
			itemTypes = append(itemTypes, item.String())
		}
		es = append(es, `i.item_type IN ('`+strings.Join(itemTypes, "', '")+`')`)
	}
	if ps.Approval != nil {
		apps := []string{}
		for _, app := range *ps.Approval {
			apps = append(apps, app.String())
		}
		es = append(es, `i.approval IN ('`+strings.Join(apps, "', '")+`')`)
	}
	if ps.Owner != nil && len(ps.Owner) != 0 {
		es = append(es, `i.owner IN ('`+strings.Join(ps.Owner, "', '")+`')`)
	}
	if ps.Tag != nil {
		es = append(es, `'`+*ps.Tag+`' = ANY(i.tags)`)
	}
	if len(es) == 0 {
		return `TRUE`
	}
	return strings.Join(es, ` AND `)
}

func paramsOrderBy(ps *SearchParams) string {
	if ps.SortBy == nil {
		return ` ORDER BY i.created_at DESC `
	}

	switch ps.SortBy.Field {
	case "rank", "upvotes":
		/* there is no prefix 'is.' because of coalesce. it removes alias from columns */
		return " ORDER BY " + ps.SortBy.Field + " " + ps.SortBy.Order + " "
	case "created_at":
		return " ORDER BY i." + ps.SortBy.Field + " " + ps.SortBy.Order + " "
	default:
		return ` ORDER BY i.created_at DESC `
	}

}

// ItemIDs gets all the existing item IDs.
func (pg *PGClient) ItemIDs() ([]string, error) {
	ids := []string{}
	err := pg.DB.Select(&ids, `SELECT id FROM items`)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// ItemNames gets the names of a list of items identified by their IDs.
func (pg *PGClient) ItemNames(ids []string) (map[string]string, error) {
	query, args, err := sqlx.In(qItemNames, ids)
	if err != nil {
		return nil, err
	}
	query = pg.DB.Rebind(query)
	var results []struct {
		ID   string `db:"id"`
		Name string `db:"name"`
	}
	err = pg.DB.Select(&results, query, args...)
	if err != nil {
		return nil, err
	}
	retval := map[string]string{}
	for _, r := range results {
		retval[r.ID] = r.Name
	}
	return retval, nil
}

const qItemNames = `SELECT id, name FROM items WHERE id IN (?)`

// CreateItem creates a new item.
func (pg *PGClient) CreateItem(item *model.Item) error {
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

	// Generate new item ID and initial attempt at a slug (which we
	// might have to change to make it unqiue).
	item.ID = chassis.NewID(model.ItemTypeIDPrefixes[item.ItemType])
	item.Slug = slug.Make(item.Name)
	if len(item.Slug) < 8 {
		item.Slug += "-" + item.ItemType.String()
	}

	// Repeatedly try to insert the item, dealing with potential slug
	// collisions by adding a random string to the slug if the insert
	// fails.
	for {
		rows, err := tx.NamedQuery(qCreateItem, item)
		if err != nil {
			return err
		}
		if rows.Next() {
			err = rows.Scan(&item.CreatedAt)
			if err != nil {
				return err
			}
			break
		}

		// Slug collision: try again...
		item.Slug = slug.Make(item.Name + " " + chassis.NewBareID(4))
	}

	return err
}

const qCreateItem = `
INSERT INTO
  items (id, item_type, slug, lang, name, description,
         featured_picture, pictures, tags, urls, attrs,
         approval, creator, owner, ownership)
 VALUES (:id, :item_type, :slug, :lang, :name, :description,
         :featured_picture, :pictures, :tags, :urls, :attrs,
         :approval, :creator, :owner, :ownership)
 ON CONFLICT DO NOTHING
 RETURNING created_at`

// UpdateItem updates the item's details in the database. The id,
// item_type, slug, approval, creator, owner, ownership and created_at
// fields are read-only using this method.
func (pg *PGClient) UpdateItem(item *model.Item, allowedOwner []string) error {
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

	check := &model.Item{}
	err = tx.Get(check, qItemBy+`i.id = $1`, item.ID)
	if err == sql.ErrNoRows {
		return ErrItemNotFound
	}
	if err != nil {
		return err
	}

	// Check read-only fields.
	if item.ItemType != check.ItemType || item.Slug != check.Slug ||
		item.Approval != check.Approval || item.Creator != check.Creator ||
		item.Owner != check.Owner || item.Ownership != check.Ownership ||
		item.CreatedAt != check.CreatedAt {
		return ErrReadOnlyField
	}

	// Check owner.
	if len(allowedOwner) > 0 {
		allowed := false
		for _, o := range allowedOwner {
			if item.Owner == o {
				allowed = true
				break
			}
		}
		if !allowed {
			return ErrItemNotOwned
		}
	}

	// Update slug if the name has changed.
	if item.Name != check.Name {
		item.Slug = slug.Make(item.Name)
	}

	// Do the update: repeatedly try to update the item, dealing with
	// potential slug collisions by adding a random string to the slug
	// if the update fails.
	for {
		result, err := tx.NamedExec(qUpdateItem, item)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 1 {
			break
		}

		// Slug collision: try again...
		item.Slug = slug.Make(item.Name + " " + chassis.NewBareID(4))
	}
	return err
}

const qUpdateItem = `
UPDATE items
 SET slug=:slug, lang=:lang, name=:name, description=:description,
     featured_picture=:featured_picture, pictures=:pictures,
     tags=:tags, urls=:urls, attrs=:attrs
 WHERE id = :id `

// UpdateItemApproval updates an item's approval state in the
// database.
func (pg *PGClient) UpdateItemApproval(id string, approval types.ApprovalState) error {
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

	result, err := tx.Exec(qUpdateApproval, id, approval)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrItemNotFound
	}

	return err
}

const qUpdateApproval = `UPDATE items SET approval=$2 WHERE id = $1`

// UpdateItemOwnership updates an item's ownership state in the
// database.
func (pg *PGClient) UpdateItemOwnership(id string,
	owner string, ownership types.OwnershipStatus) error {
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

	result, err := tx.Exec(qUpdateOwnership, id, owner, ownership)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrItemNotFound
	}

	return err
}

const qUpdateOwnership = `UPDATE items SET owner=$2, ownership=$3 WHERE id = $1`

// DeleteItem deletes the given item.
// TODO: ADD SOME SORT OF ARCHIVAL MECHANISM INSTEAD.
func (pg *PGClient) DeleteItem(id string, allowedOwner []string) ([]string, error) {
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

	// Check that item exists, so that we can distinguish between bad
	// item ID and bad owner below.
	check, err := pg.ItemByID(id)
	if err != nil {
		return nil, ErrItemNotFound
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
			return nil, ErrItemNotOwned
		}
	}

	// Try to delete the item.
	result, err := tx.Exec(qDeleteItem, id)
	if err != nil {
		return nil, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows != 1 {
		return nil, ErrItemNotOwned
	}
	return check.Pictures, err
}

const qDeleteItem = `
DELETE FROM items WHERE id = $1`

// TagsForUser returns all the tags used in items owned by a user.
func (pg *PGClient) TagsForUser(userID string) ([]string, error) {
	tags := []string{}
	err := pg.DB.Select(&tags, qTagsForUser, userID)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

const qTagsForUser = `
SELECT DISTINCT unnest(tags) AS tag
  FROM items WHERE owner = $1 ORDER BY tag`

func intersectIDs(ids1 []string, ids2 []string) []string {
	if len(ids1) == 0 {
		return ids2
	}

	if len(ids2) == 0 {
		return ids1
	}

	ids := map[string]bool{}
	for _, id := range ids1 {
		ids[id] = true
	}

	result := []string{}
	for _, id := range ids2 {
		if ids[id] {
			result = append(result, id)
		}
	}
	return result
}

func makeOrderMap(ids []string) map[string]int {
	res := map[string]int{}
	for i, id := range ids {
		res[id] = i
	}
	return res
}

// UpdateAvailability updates the item's availability. More precisely, the attrs column.
// The other fields won't be affected because the Patch method is not used. item.UpdateAvailability
// is used instead.
func (pg *PGClient) UpdateAvailability(item *model.Item) error {
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

	check := &model.Item{}
	err = tx.Get(check, qItemBy+`i.id = $1`, item.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrItemNotFound
		}
		return err
	}

	result, err := tx.NamedExec(qUpdateItemAvailability, item)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return err
}

const qUpdateItemAvailability = `
UPDATE items
 SET attrs=:attrs
 WHERE id = :id`
