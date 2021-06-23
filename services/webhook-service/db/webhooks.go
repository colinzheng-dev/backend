package db

import (
	"database/sql"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/webhook-service/model"
)

func (pg *PGClient) WebhooksByOwner(owner string) (*[]model.Webhook, error) {
	hooks := []model.Webhook{}
	if err := pg.DB.Select(&hooks, qGetWebhooks + ` owner = $1 `, owner); err != nil {
		return nil, err
	}
	return &hooks, nil
}


func (pg *PGClient) WebhookByOwnerAndEventType(owner, eventType string) (*model.Webhook, error) {
	hooks := model.Webhook{}
	if err := pg.DB.Get(&hooks, qGetWebhooks + ` owner = $1 AND ($2=ANY(events) OR '*'=ANY(events)) `, owner, eventType); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrWebhookNotFound
		}
		return nil, err
	}
	return &hooks, nil
}

func (pg *PGClient) IsEventHandled(owner, eventType string) (*bool, error) {
	var result bool

	if err := pg.DB.QueryRow(qIsEventHandled,  owner, eventType).Scan(&result); err != nil &&  err != sql.ErrNoRows {
		return nil, err
	}
	return &result, nil
}


const qIsEventHandled = `
	SELECT exists(
	    SELECT id
		FROM webhooks
		WHERE owner = $1 AND ($2=ANY(events) OR '*'=ANY(events))
	)
`

func (pg *PGClient) WebhookByID(hookID string) (*model.Webhook, error) {
	hook := model.Webhook{}
	if err := pg.DB.Get(&hook, qGetWebhooks + ` id = $1`,hookID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrWebhookNotFound
		}
		return nil, err
	}
	return &hook, nil
}

const qGetWebhooks = `
	SELECT id, owner, url, enabled, livemode, events, secret, created_at
	FROM webhooks
	WHERE 
`

// CreateWebhook creates a new webhook entry.
func (pg *PGClient) CreateWebhook(wh *model.Webhook) error {
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

	wh.ID = chassis.NewID("web")
	wh.Secret = chassis.GenerateUUID("whk")
	rows, err := tx.NamedQuery(qCreateWebhook, wh)
	if err != nil {
		return err
	}
	if rows.Next() {
		err = rows.Scan(&wh.CreatedAt)
		if err != nil {
			return err
		}
	}

	return err
}

const qCreateWebhook = `
INSERT INTO webhooks (id, owner, url, enabled, livemode, events, secret)
VALUES (:id, :owner, :url, :enabled, :livemode, :events, :secret)
ON CONFLICT DO NOTHING
RETURNING created_at`

// DeleteWebhook deletes a webhook entry with a certain id
func (pg *PGClient) DeleteWebhook(hookID string) error {
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

	// Try to delete the cart item.
	result, err := tx.Exec(qDeleteWebhook, hookID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrWebhookNotFound //TODO: CHANGE ERROR
	}
	return err
}

const qDeleteWebhook = ` DELETE FROM webhooks WHERE id = $1 `


func (pg *PGClient) UpdateWebhook(hook *model.Webhook) error {
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

	//check if cart item exists
	check := &model.Webhook{}

	if err = tx.Get(check, qGetWebhooks + ` id = $1`, hook.ID); err != nil {
		if err == sql.ErrNoRows {
			return ErrWebhookNotFound
		}
		return err
	}

	result, err := tx.NamedExec(qUpdateWebhook, hook)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return err
}

const qUpdateWebhook = `
UPDATE webhooks
SET url=:url, enabled=:enabled, livemode=:livemode, events=:events
WHERE id = :id `
