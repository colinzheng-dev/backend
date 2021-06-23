package db

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/veganbase/backend/services/purchase-service/model"
)

const qSubscriptionPurchasesProcessingBy = `
SELECT 
	id, reference, is_processing_day, status, created_at, started_at, ended_at
FROM subscription_purchases_processing WHERE `

// SubscriptionPurchaseProcessingByID looks up a subscription item by its id.
func (pg *PGClient) SubscriptionPurchaseProcessingByID(id string) (*model.SubscriptionPurchaseProcessing, error) {
	processing := &model.SubscriptionPurchaseProcessing{}

	if err := sqlx.Get(pg.DB, processing, qSubscriptionPurchasesProcessingBy + `id = $1`, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSubscriptionPurchaseProcessingNotFound
		}
		return nil, err
	}
	return processing, nil
}

// SubscriptionPurchaseProcessingByReference looks up a subscription processing by its reference date.
func (pg *PGClient) SubscriptionPurchaseProcessingByReference(ref string) (*[]model.SubscriptionPurchaseProcessing, error) {
	processing := &[]model.SubscriptionPurchaseProcessing{}

	if err := sqlx.Select(pg.DB, processing, qSubscriptionPurchasesProcessingBy + `reference = $1`, ref); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSubscriptionPurchaseProcessingNotFound
		}
		return nil, err
	}
	return processing, nil
}

// UpdateSubscriptionItemsProcessing updates a subscription item processing.
func (pg *PGClient) UpdateSubscriptionPurchaseProcessing(sub *model.SubscriptionPurchaseProcessing) error {
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

	check := &model.SubscriptionPurchaseProcessing{}
	if err = tx.Get(check, qSubscriptionPurchasesProcessingBy + `id = $1`, sub.ID); err != nil {
		if err == sql.ErrNoRows {
			return ErrSubscriptionPurchaseProcessingNotFound
		}
		return err
	}

	result, err := tx.NamedExec(qUpdateSubscriptionPurchasesProcessing, sub)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return err
}

const qUpdateSubscriptionPurchasesProcessing = `
UPDATE subscription_purchases_processing
SET status = :status,
    is_processing_day = :is_processing_day,
    started_at = :started_at, 
    ended_at = :ended_at
WHERE id = :id`
