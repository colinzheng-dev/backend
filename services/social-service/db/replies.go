package db

import (
	"database/sql"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/social-service/model"
)


const qReplyBy = `
	SELECT id, parent_id, owner, is_edited, is_deleted, pictures, attrs, created_at
	FROM replies
	 `
// ReplyById returns a reply with a certain id
func (pg *PGClient) ReplyById(replyId string) (*model.Reply, error) {
	reply:= model.Reply{}
	if err := pg.DB.Get(&reply, qReplyBy + " WHERE id = $1", replyId); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrReplyNotFound
		}
		return nil, err
	}
	return &reply, nil
}

// RepliesByParentId returns all replies a parent (a post or other reply) received.
func (pg *PGClient) RepliesByParentId(parentId string) (*[]model.Reply, error) {
	replies:= []model.Reply{}
	if err := pg.DB.Select(&replies, qReplyBy + " WHERE parent_id = $1 ORDER BY created_at ASC", parentId); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &replies, nil
}


const qCreateReply = `
INSERT INTO
	replies (id, parent_id, owner, is_edited, is_deleted, pictures, attrs)
VALUES (:id, :parent_id, :owner, :is_edited, :is_deleted, :pictures, :attrs)
ON CONFLICT DO NOTHING
RETURNING created_at
`

// CreateReply creates a new reply
func (pg *PGClient) CreateReply(reply *model.Reply) error {
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
	reply.Id = chassis.NewID("rpl")
	rows, err := tx.NamedQuery(qCreateReply, reply)
	if err != nil {
		return err
	}
	if rows.Next() {
		err = rows.Scan(&reply.CreatedAt)
		if err != nil {
			return err
		}
	}
	return err
}


const qUpdateReply = `
UPDATE replies
 SET is_edited=:is_edited, is_deleted=:is_deleted,
      pictures=:pictures, attrs=:attrs
 WHERE id = :id`

// UpdateReply updates the reply details in the database.
func (pg *PGClient) UpdateReply(reply *model.Reply) error {
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

	check := &model.Reply{}
	if err = tx.Get(check, qReplyBy +` WHERE id = $1`, reply.Id); err != nil {
		if err == sql.ErrNoRows {
			return ErrReplyNotFound
		}
		return err
	}

	// Check read-only fields.
	if reply.ParentId != check.ParentId || 	reply.Owner != check.Owner || reply.CreatedAt != check.CreatedAt {
		return ErrReadOnlyField
	}

	result, err := tx.NamedExec(qUpdateReply, reply)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrReplyNotFound
	}

	return err
}


const qDeleteReply = `
DELETE FROM replies
WHERE id = $1
`

const qMarkReplyForDeletion = `
UPDATE replies
SET is_deleted = true
WHERE id = $1
`
// DeletePost removes a post
func (pg *PGClient) DeleteReply(replyId string) error {
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
	//TODO: CHECK IF EXISTS NESTED REPLIES. IF NEGATIVE, DELETES IT. OTHERWISE MARK FOR DELETION.
	result, err := tx.Exec(qDeleteReply, replyId)
	// Try to delete the cart item.
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrReplyNotFound
	}
	return err
}

