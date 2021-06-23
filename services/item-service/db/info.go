package db

import (
	"github.com/jmoiron/sqlx"
	"github.com/veganbase/backend/services/item-service/model"
)

// GetItemTypeInfo looks up quantity of each item_type on the database.
func (pg *PGClient) GetItemTypeInfo() (*[]model.ItemTypeInfo, error) {

	itemTypeInfo := &[]model.ItemTypeInfo{}
	if err := sqlx.Select(pg.DB, itemTypeInfo, qGetItemQuantities); err != nil {
		return nil, err
	}

	return itemTypeInfo, nil
}


const qGetItemQuantities = `
SELECT item_type, quantity FROM (
	SELECT item_type, count(*) AS quantity FROM ITEMS GROUP BY item_type
	UNION ALL
	SELECT 'all', count(*) as quantity FROM items
) AS r ORDER BY r.item_type ASC `


const qItemInfo = `
SELECT id, slug, name FROM items WHERE id IN (?) `

// Info gets minimal information about a list of items given their IDs.
func (pg *PGClient) Info(ids []string) (map[string]model.Info, error) {
	retval := map[string]model.Info{}

	query, queryArgs, err := sqlx.In(qItemInfo, ids)
	if err != nil {
		return nil, err
	}
	query = pg.DB.Rebind(query)
	result := []model.Info{}

	if err = pg.DB.Select(&result, query, queryArgs...); err != nil {
		return nil, err
	}

	for _, r := range result {
		retval[r.ID] = r
	}

	return retval, nil
}
