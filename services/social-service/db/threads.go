package db

import (
	"database/sql"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/social-service/model"

)


const qThreadBy = `
	SELECT id, subject, author, content, attachments, lock_reply, participants, status, is_edited, created_at
	FROM threads `

// PostByUser returns a thread by its id.
func (pg *PGClient) ThreadByID(threadID string) (*model.Thread, error) {
	tr:= model.Thread{}
	if err := pg.DB.Get(&tr, qThreadBy + " WHERE id = $1", threadID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrThreadNotFound
		}
		return nil, err
	}
	return &tr, nil
}

func (pg *PGClient) GetThreads(params *DatabaseParams) (*[]model.Thread, *uint, error) {
	results := []model.Thread{}
	var total uint
	q := qThreadBy + paramsWhere(params) + paramsOrderBy(params)
	if params.Pagination != nil {
		q += chassis.Paginate(params.Pagination.Page, params.Pagination.PerPage)
	}

	if err := pg.DB.Select(&results, q); err != nil {
		return nil, &total, err
	}
	if err := chassis.GetTotalResultsFromQuery(q, &total, pg.DB); err != nil {
		return nil, nil, err
	}

	return &results, &total, nil
}


const qCreateThread = `
INSERT INTO
	threads(id, subject, author, content, attachments, lock_reply, participants, status, is_edited)
VALUES (:id, :subject, :author, :content, :attachments, :lock_reply, :participants, 'open', :is_edited)
ON CONFLICT DO NOTHING
RETURNING created_at
`

// CreateThread creates a new thread
func (pg *PGClient) CreateThread(th *model.Thread) error {
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
	th.ID = chassis.NewID("thr")

	rows, err := tx.NamedQuery(qCreateThread, th)
	if err != nil {
		return err
	}
	if rows.Next() {
		err = rows.Scan(&th.CreatedAt)
		if err != nil {
			return err
		}
	}
	return err
}

const qUpdateThread = `
UPDATE threads
 SET is_edited=:is_edited, attachments =:attachments, content=:content
 WHERE id = :id`

// UpdateThread updates the thread details in the database.
func (pg *PGClient) UpdateThread(th *model.Thread) error {
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

	check := &model.Thread{}
	if err = tx.Get(check, qThreadBy + ` WHERE id = $1`, th.ID); err != nil {
		if err == sql.ErrNoRows {
			return ErrThreadNotFound
		}
		return err
	}
	th.IsEdited = true

	result, err := tx.NamedExec(qUpdateThread, th)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrThreadNotFound
	}

	return err
}

const qThreadChangeStatus = `
UPDATE threads 
SET status = $1
WHERE id = $2
`
// ChangeThreadStatus change the status of a thread
func (pg *PGClient) ChangeThreadStatus(threadID, status string) error {
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

	result, err := tx.Exec(qThreadChangeStatus, status, threadID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrThreadNotFound
	}
	return err
}
