package db

import (
	"database/sql"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/social-service/model"
	"strings"
)

const qUpvoteBy = `
	SELECT upvote_id, item_id, user_id, created_at
	FROM upvotes
	WHERE `

// UpvoteByUserAndItemId returns an upvote of a certain user for a specific item if exists.
func (pg *PGClient) UpvoteByUserAndItemId(userId, itemId string) (*model.Upvote, error) {
	upvote:= model.Upvote{}
	if err := pg.DB.Get(&upvote, qUpvoteBy + "user_id = $1 and item_id = $2", userId, itemId); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUpvoteNotFound
		}
		return nil, err
	}
	return &upvote, nil
}

const qItemsUpvotedByUserId = `
	SELECT item_id
	FROM upvotes
	WHERE user_id = $1 `

// ListUserUpvotes returns an upvote of a certain user.
func (pg *PGClient) ListUserUpvotes(userId string) (*[]string, error) {
	upvotes:= []string{}
	if err := pg.DB.Select(&upvotes, qItemsUpvotedByUserId , userId); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &upvotes, nil
}


const qCreateUpvote = `
INSERT INTO
	upvotes (upvote_id, user_id, item_id)
VALUES (:upvote_id, :user_id, :item_id)
ON CONFLICT DO NOTHING
RETURNING created_at
`

// CreateLove associates the given userID with a item_id
func (pg *PGClient) CreateUpvote(upvote *model.Upvote) error {
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

	upvote.Id = chassis.NewID("upv")
	rows, err := tx.NamedQuery(qCreateUpvote, upvote)
	if err != nil {
		return err
	}
	if !rows.Next() {
		return ErrAlreadyUpvoted
	}
	return rows.Scan(&upvote.CreatedAt)
}

const qListUserUpvotes = `
	SELECT item_id
	FROM upvotes
	WHERE user_id = $1
`

const qUpvoteQuantity = `
	SELECT item_id, count(upvote_id) as quantity
	FROM upvotes
	 `

// qListUserUpvotes returns a list of item_ids that a user likes.
func (pg *PGClient) UpvoteQuantityByItemId(ids []string) (*[]model.UpvoteQuantityInfo, error) {
	result := []model.UpvoteQuantityInfo{}
	whereClause := ""
	if len(ids) > 0 {
		whereClause = "WHERE item_id IN ('"+ strings.Join(ids, "','") + "')"
	}
	err := pg.DB.Select(&result, qUpvoteQuantity + whereClause + " GROUP BY item_id")

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return &result, nil
}

const qDeleteUpvote = `
DELETE FROM upvotes
WHERE upvote_id = $1
`
// DeleteUserSubscription removes a subscriptionID from the userID
func (pg *PGClient) DeleteUpvote(upvoteId string) error {
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

	result, err := tx.Exec(qDeleteUpvote, upvoteId )
	// Try to delete the cart item.
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrUpvoteNotFound
	}
	return err
}


const qListWhoUpvoted = `
SELECT user_id
FROM upvotes
WHERE item_id = $1
`

// ListWhoUpvoted returns a list of userIDs that upvoted an item
func (pg *PGClient) ListWhoUpvoted(itemId string) ([]string, error) {
	users := []string{}
	err := pg.DB.Select(&users, qListWhoUpvoted, itemId)
	if err != sql.ErrNoRows && err != nil {
		return nil, err
	}
	return users, nil
}
