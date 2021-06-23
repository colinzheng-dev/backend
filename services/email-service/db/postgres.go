package db

import (
	"context"

	"github.com/jmoiron/sqlx"

	// Import Postgres DB driver.
	_ "github.com/lib/pq"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/email-service/model"
)

// PGClient is a wrapper for the email database connection.
type PGClient struct {
	DB *sqlx.DB
}

// NewPGClient creates a new email database connection.
func NewPGClient(ctx context.Context, dbURL string) (*PGClient, error) {
	db, err := chassis.DBConnect(ctx, "email", dbURL, Asset, AssetDir)
	if err != nil {
		return nil, err
	}
	return &PGClient{db}, nil
}

// Topics retrieves the list of registered topics.
func (pg *PGClient) Topics() ([]model.Topic, error) {
	topics := []model.Topic{}
	err := pg.DB.Select(&topics, `SELECT id, name, send_address, created_at FROM topics`)
	if err != nil {
		return nil, err
	}
	return topics, nil
}

// SaveEvent saves an event to the database.
func (pg *PGClient) SaveEvent(label string, eventData interface{}, inTx func() error) error {
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
	err = chassis.SaveEvent(tx, label, eventData, inTx)
	return err
}
