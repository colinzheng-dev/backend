package db

import (
	"context"

	"github.com/jmoiron/sqlx"

	// Import Postgres DB driver.
	_ "github.com/lib/pq"

	"github.com/veganbase/backend/chassis"
)

// PGClient is a wrapper for the user database connection.
type PGClient struct {
	DB *sqlx.DB
}

// NewPGClient creates a new user database connection.
func NewPGClient(ctx context.Context, dbURL string) (*PGClient, error) {
	db, err := chassis.DBConnect(ctx, "user", dbURL, Asset, AssetDir)
	if err != nil {
		return nil, err
	}
	return &PGClient{db}, nil
}

// SaveEvent saves an event to the database.
func (pg *PGClient) SaveEvent(topic string, eventData interface{}, inTx func() error) error {
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
	err = chassis.SaveEvent(tx, topic, eventData, inTx)
	return err
}
