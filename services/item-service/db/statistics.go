package db

import (
	"database/sql"
	"database/sql/driver"
)

const qCheckStatisticExistanceBy = `
SELECT 1
  FROM item_statistics WHERE `

const qGetStatisticBy = `
SELECT item_id, rank, upvotes
  FROM item_statistics WHERE `

const qInsertStatisticRecord = `
INSERT INTO item_statistics 
(item_id, rank, upvotes)
VALUES ($1, $2, $3) `

const qUpdateRankStatistic = `
UPDATE item_statistics SET rank = $1
WHERE item_id = $2 `

const qUpdateUpvoteStatistic = `
UPDATE item_statistics SET upvotes = $1
WHERE item_id = $2 `

func (pg *PGClient) SetItemRank(itemId string, rank float64) error {
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
	var check int
	if err := pg.DB.Get(&check, qCheckStatisticExistanceBy+`item_id = $1`, itemId); err != nil && err != sql.ErrNoRows {
		return err
	}

	var result driver.Result
	if check == 1 {
		result, err = tx.Exec(qUpdateRankStatistic, rank, itemId)
	} else {
		result, err = tx.Exec(qInsertStatisticRecord, itemId, rank, 0)
	}

	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrStatisticNotFound
	}

	return err
}

func (pg *PGClient) SetItemUpvotes(itemId string, upvotes int) error {
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
	var check struct {
		ItemID  string  `db:"item_id"`
		Rank    float64 `db:"rank"`
		Upvotes int     `db:"upvotes"`
	}
	if err := pg.DB.Get(&check, qGetStatisticBy+`item_id = $1`, itemId); err != nil && err != sql.ErrNoRows {
		return err
	}
	if check.Upvotes == upvotes {
		return nil
	}
	var result driver.Result
	if check.ItemID != "" {
		result, err = tx.Exec(qUpdateUpvoteStatistic, upvotes, itemId)
	} else {
		result, err = tx.Exec(qInsertStatisticRecord, itemId, 0, upvotes) //rank is zero
	}

	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrStatisticNotFound
	}

	return err
}
