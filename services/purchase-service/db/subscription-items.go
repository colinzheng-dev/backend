package db

import (
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/purchase-service/model"
	"github.com/veganbase/backend/services/purchase-service/model/types"
	"time"
)

const qSubscriptionItemBy = `
SELECT 
	id, owner, item_id, item_type, address_id, origin, quantity, other_info, delivery_every, 
	status, next_delivery, last_delivery, created_at, active_since, paused_since, deleted_at
FROM subscription_items WHERE `

// SubscriptionItemById looks up a subscription item by its ID.
func (pg *PGClient) SubscriptionItemById(subID string) (*model.SubscriptionItem, error) {
	sub := &model.SubscriptionItem{}

	if err := sqlx.Get(pg.DB, sub, qSubscriptionItemBy + `id = $1`, subID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSubscriptionItemNotFound
		}
		return nil, err
	}
	return sub, nil
}

// SubscriptionItemsByOwner retrieves all subscription items of a specific owner
func (pg *PGClient) SubscriptionItemsByOwner(ownerID string, params *chassis.Pagination) (*[]model.SubscriptionItem, *uint, error) {
	subs := &[]model.SubscriptionItem{}
	var total uint

	q := qSubscriptionItemBy + `owner = '` + ownerID + `' ORDER BY created_at DESC `
	if params != nil {
		q += chassis.Paginate(params.Page, params.PerPage)
	}

	if err := sqlx.Select(pg.DB, subs, q); err != nil && err != sql.ErrNoRows {
		return nil, nil,  err
	}
	if err := chassis.GetTotalResultsFromQuery(q, &total, pg.DB); err != nil {
		return nil, nil, err
	}
	return subs, &total, nil
}

func (pg *PGClient) SubscriptionItemsByOwnerAndReference(ownerID string, reference string) (*[]model.SubscriptionItem, error) {
	subs := &[]model.SubscriptionItem{}

	q := qSubscriptionItemBy + ` owner = $1 AND status = 'active' AND EXTRACT(MONTH FROM to_date($2,'YYYY/MM/DD')) = next_delivery`

	if err := sqlx.Select(pg.DB, subs, q, ownerID, reference); err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return subs,  nil
}

// CreateSubscription creates a new subscription item.
func (pg *PGClient) CreateSubscription(sub *model.SubscriptionItem) error {
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
	//creating a new subscription id
	sub.ID = chassis.NewID("sub")

	//inserting the subscription
	rows, err := tx.NamedQuery(qCreateSubscriptionItem, sub)
	if err != nil {
		return err
	}
	if rows.Next() {
		if err = rows.Scan(&sub.CreatedAt, &sub.ActiveSince); err != nil {
			return err
		}
		if err = rows.Close(); err != nil {
			return err
		}
	}

	return err
}

const qCreateSubscriptionItem = `
INSERT INTO
	subscription_items (id, owner, item_id, item_type, address_id,
	                    origin, quantity, other_info, delivery_every, 
						status, next_delivery, last_delivery,  created_at, active_since)
VALUES (:id, :owner, :item_id, :item_type, :address_id,
        :origin, :quantity, :other_info, :delivery_every, 
		:status, :next_delivery, now(),  now(), now())
ON CONFLICT DO NOTHING
RETURNING created_at, active_since`

// UpdateSubscriptionItem updates a subscription item quantity, next_delivery and delivery_every.
func (pg *PGClient) UpdateSubscriptionItem(sub *model.SubscriptionItem) error {
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

	check := &model.SubscriptionItem{}
	if err = tx.Get(check, qSubscriptionItemBy + `id = $1`, sub.ID); err != nil {
		if err == sql.ErrNoRows {
			return ErrSubscriptionItemNotFound
		}
		return err
	}

	result, err := tx.NamedExec(qUpdateSubscriptionItem, sub)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return err
}

const qUpdateSubscriptionItem = `
UPDATE subscription_items
SET address_id = :address_id,
    quantity = :quantity, 
    next_delivery = :next_delivery,
    delivery_every = :delivery_every
WHERE id = :id`



// FlipActivationState sets a subscription item state to active or paused.
func (pg *PGClient) FlipActivationState(sub *model.SubscriptionItem) error {
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


	var result sql.Result
	now := time.Now()
	switch sub.Status {
	case types.Active:
		sub.Status = types.Paused
		sub.NextDelivery = 0
		sub.PausedSince = &now
		sub.ActiveSince = nil

	case types.Paused:
		sub.Status = types.Active
		if sub.NextDelivery < int(time.Now().Month()) {
			sub.NextDelivery = (int(time.Now().Month()) + sub.DeliveryEvery) % 12
		}
		sub.PausedSince = nil
		sub.ActiveSince = &now
	default:
		return errors.New("cannot change state from "+ sub.Status.String())
	}
	result, err = tx.NamedExec(qFlipState, sub)

	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return err
}


const qFlipState = `
UPDATE subscription_items
SET status = :status,
    next_delivery = :next_delivery,
    active_since = :active_since,
    paused_since = :paused_since
WHERE id = :id `


// DeleteSubscriptionItem sets a subscription item state as deleted.
func (pg *PGClient) DeleteSubscriptionItem(sub *model.SubscriptionItem) error {
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

	if sub.Status == types.Deleted {
		return nil
	}

	now := time.Now()
	sub.ActiveSince = nil
	sub.PausedSince = nil
	sub.DeletedAt = &now
	sub.Status = types.Deleted

	result, err := tx.NamedExec(qDeleteSubscriptionItem, sub)

	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return err
}

const qDeleteSubscriptionItem = `
UPDATE subscription_items
SET status = 'deleted',
    active_since = :active_since,
    paused_since = :paused_since,
	deleted_at = :deleted_at
WHERE id = :id `



// UpdateLastDelivery updates last delivery field.
func (pg *PGClient) UpdateLastDelivery(sub *model.SubscriptionItem) error {
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

	check := &model.SubscriptionItem{}
	if err = tx.Get(check, qSubscriptionItemBy + `id = $1`, sub.ID); err != nil {
		if err == sql.ErrNoRows {
			return ErrSubscriptionItemNotFound
		}
		return err
	}

	result, err := tx.NamedExec(qUpdateLastDelivery, sub)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return err
}

const qUpdateLastDelivery = `
UPDATE subscription_items
SET last_delivery = now()
WHERE id = :id`
