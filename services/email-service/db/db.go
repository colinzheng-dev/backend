package db

import (
	"github.com/veganbase/backend/services/email-service/model"
)

// DB describes the database operations used by the email service.
type DB interface {
	// Topics gets the list of registered topics.
	Topics() ([]model.Topic, error)

	// SaveEvent saves an event to the database.
	SaveEvent(label string, eventData interface{}, inTx func() error) error
}

//go:generate go-bindata -pkg db -o migrations.go migrations/...
