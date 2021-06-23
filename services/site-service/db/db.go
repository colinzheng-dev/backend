package db

import (
	"github.com/veganbase/backend/services/site-service/model"
)

// DB describes the database operations used by the user service.
type DB interface {
	// Sites gets the list of sites.
	Sites() (map[string]*model.Site, error)

	// SaveEvent saves an event to the database.
	SaveEvent(label string, eventData interface{}, inTx func() error) error
}

//go:generate go-bindata -pkg db -o migrations.go migrations/...
