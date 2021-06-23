package db

import (
	"database/sql"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/social-service/model"
)

const qMessageBy = `
	SELECT id, parent_id, author, content, attachments, is_edited, is_deleted, created_at
	FROM messages
	 `
// MessageByID returns a message with a certain id
func (pg *PGClient) MessageByID(messageID string) (*model.Message, error) {
	msg:= model.Message{}
	if err := pg.DB.Get(&msg, qMessageBy + " WHERE id = $1", messageID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrMessageNotFound
		}
		return nil, err
	}
	return &msg, nil
}

// MessagesByParentID returns all messages a parent (a thread) contains.
func (pg *PGClient) MessagesByParentID(parentId string) (*[]model.Message, error) {
	msgs:= []model.Message{}
	if err := pg.DB.Select(&msgs, qMessageBy + " WHERE parent_id = $1 ORDER BY created_at ASC", parentId); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &msgs, nil
}

const qCreateMessage = `
INSERT INTO
	messages (id, parent_id, author, content, attachments, is_edited, is_deleted)
VALUES (:id, :parent_id, :author, :content, :attachments, :is_edited, :is_deleted)
ON CONFLICT DO NOTHING
RETURNING created_at
`

// CreateMessage creates a new message
func (pg *PGClient) CreateMessage(msg *model.Message) error {
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
	msg.ID = chassis.NewID("msg")
	rows, err := tx.NamedQuery(qCreateMessage, msg)
	if err != nil {
		return err
	}
	if rows.Next() {
		err = rows.Scan(&msg.CreatedAt)
		if err != nil {
			return err
		}
	}
	return err
}


const qUpdateMessage = `
UPDATE messages
 SET content=:content, is_edited=:is_edited,
      attachments=:attachments
 WHERE id = :id`

// UpdateMessage updates the message details in the database.
func (pg *PGClient) UpdateMessage(msg *model.Message) error {
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

	check := &model.Message{}
	if err = tx.Get(check, qMessageBy +` WHERE id = $1`, msg.ID); err != nil {
		if err == sql.ErrNoRows {
			return ErrMessageNotFound
		}
		return err
	}


	msg.IsEdited = true

	result, err := tx.NamedExec(qUpdateMessage, msg)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrMessageNotFound
	}

	return err
}


const qMarkMessageAsDeleted = `UPDATE messages SET is_deleted = true WHERE id = $1 `
// DeleteMessage marks a message as deleted
func (pg *PGClient) DeleteMessage(msgID string) error {
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
	result, err := tx.Exec(qMarkMessageAsDeleted, msgID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrMessageNotFound
	}
	return err
}

const qGetAuthorsBy = `
	SELECT DISTINCT m.author
	FROM messages m 
	INNER JOIN threads t ON t.id = m.parent_id 
	WHERE t.id IN `

// MessagesByParentID returns all messages a parent (a thread) contains.
func (pg *PGClient) AuthorsByThreadID(ids []string) ([]string, error) {
	authors:= []string{}
	if err := pg.DB.Select(&authors, qGetAuthorsBy + chassis.BuildInArgument(ids)); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return authors, nil
}

