package pubsub

import "errors"

// ErrUnknownTopic is the error returned when Publish or Subscribe is
// called on an unknown topic: topics must be set up by calling
// EnsureTopic before use.
var ErrUnknownTopic = errors.New("unknown topic")

// SubscriptionMode is used to distinguish between two distinct types
// of subscription to a pub/sub topic.
type SubscriptionMode int

const (
	// CompetingConsumers is a subscription mode where each message on a
	// topic is delivered to a single subscriber -- it is used for
	// service-oriented subscriptions, where each message on the topic
	// is effectively a command to some subscribing service to perform
	// an action, and the action should only be performed once.
	CompetingConsumers SubscriptionMode = 0

	// Fanout is a subscription mode where each message on a topic is
	// delivered to all subscribers -- it is used for notifications that
	// should be handled by all instances of listening services.
	Fanout SubscriptionMode = 1
)

// PubSub is a simplified interface to a publish-subscribe system.
type PubSub interface {
	Publish(topic string, data interface{}) error
	Subscribe(topic, sub string, mode SubscriptionMode) (chan []byte, func(), error)
	Close()
}
