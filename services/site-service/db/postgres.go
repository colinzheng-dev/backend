package db

import (
	"context"

	"github.com/jmoiron/sqlx"

	// Import Postgres DB driver.
	_ "github.com/lib/pq"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/site-service/model"
)

// PGClient is a wrapper for the site database connection.
type PGClient struct {
	DB *sqlx.DB
}

// NewPGClient creates a new site database connection.
func NewPGClient(ctx context.Context, dbURL string) (*PGClient, error) {
	db, err := chassis.DBConnect(ctx, "site", dbURL, Asset, AssetDir)
	if err != nil {
		return nil, err
	}
	return &PGClient{db}, nil
}

// Sites retrieves the site list.
func (pg *PGClient) Sites() (map[string]*model.Site, error) {
	sites := []model.Site{}
	err := pg.DB.Select(&sites, selectSites)
	if err != nil {
		return nil, err
	}
	sitemap := map[string]*model.Site{}
	for _, s := range sites {
		s := s
		sitemap[s.ID] = &s
	}
	return sitemap, nil
}

const selectSites = `
SELECT id, name, url, email_domain, signature, fee, created_at FROM sites`

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
