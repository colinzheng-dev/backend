package db

import (
	"errors"
	"github.com/veganbase/backend/services/payment-service/model"
)

// ErrPaymentIntentNotFound is the error returned when an attempt is made to
// access or manipulate a payment intent with an unknown ID.
var ErrPaymentIntentNotFound = errors.New("intent ID not found")

// ErrReceivedEventNotFound is the error returned when an attempt is made to
// access or manipulate a payment intent with an unknown ID.
var ErrReceivedEventNotFound = errors.New("received event not found")

var ErrPendingEventNotFound = errors.New("pending event not found")

var ErrPendingTransferNotFound = errors.New("pending transfer not found")

// DB describes the database operations used by the payment service.
type DB interface {
	// PaymentIntents
	PaymentIntentsByOwner(owner string) (*[]model.PaymentIntent, error)
	SuccessfulPaymentIntentsByOrigin(origin string) (*[]model.PaymentIntent, error)
	PaymentIntentByIntentId(id string) (*model.PaymentIntent, error)
	CreatePaymentIntent(pi *model.PaymentIntent) error
	UpdatePaymentIntent(pi *model.PaymentIntent) error

	//Transfers
	TransfersByDestination(destination string) (*[]model.Transfer, error)
	TransfersByOrigin(origin string) (*[]model.Transfer, error)
	CreateTransfer(tr *model.Transfer) error
	CreateTransfers(remainder *model.TransferRemainder, origin string) error

	//Audit
	ReceivedEventByEventId(id string) (*model.ReceivedEvent, error)
	CreateReceivedEvent(event *model.ReceivedEvent) error
	UpdateReceivedEvent(event *model.ReceivedEvent) error
	CreateTransferRemainder(remainder *model.TransferRemainder) error
	CreateErrorLog(log *model.ErrorLog) error

	//Pending Queues
	PendingEvents() (*[]model.PendingEvent, error)
	PendingEventByEventID(eventId string) (*model.PendingEvent, error)
	CreatePendingEvent(pe *model.PendingEvent) error
	DeletePendingEvent(eventId string) error
	UpdatePendingEvent(pe *model.PendingEvent) error

	PendingTransfers() (*[]model.PendingTransfer, error)
	CreatePendingTransfer(pt *model.PendingTransfer) error
	DeletePendingTransfer(id int64) error

	// SaveEvent saves an event to the database.
	SaveEvent(topic string, eventData interface{}, inTx func() error) error
}

//go:generate go-bindata -pkg db -o migrations.go migrations/...
