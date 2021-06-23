package db

import (
	"errors"
	"github.com/veganbase/backend/services/webhook-service/model"
)
var ErrWebhookNotFound = errors.New("webhook not found")

var ErrEventNotFound = errors.New("event not found")

// DB describes the database operations used by the search service.
type DB interface {
	WebhooksByOwner(owner string) (*[]model.Webhook, error)
	WebhookByOwnerAndEventType(owner, eventType string) (*model.Webhook, error)
	WebhookByID(hookID string) (*model.Webhook, error)
	CreateWebhook(hook *model.Webhook) error
	DeleteWebhook(hookID string) error
	UpdateWebhook(hook *model.Webhook) error

	AddEvent(e *model.Event) error
	IsEventHandled(owner, eventType string) (*bool, error)
	EventByID(eventID string) (*model.Event, error)
	EventsByDestination(destination string) (*[]model.Event, error)
	PendingEvents() (*[]model.Event, error)
	UpdateEvent(ev *model.Event) error
	SetSentStatus(eventID string) error
	IncreaseFailedAttempts(eventID string) error
	DisableRetryFlag(eventID string) error

	SaveEvent(label string, eventData interface{}, inTx func() error) error
}

//go:generate go-bindata -pkg db -o migrations.go migrations/...
