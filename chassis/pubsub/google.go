package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	google "cloud.google.com/go/pubsub"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
)

// GooglePubSub provides an interface to Google Cloud Pub/Sub.
type GooglePubSub struct {
	mu     sync.Mutex
	ctx    context.Context
	pubsub *google.Client
	topics map[string]*google.Topic
	subs   map[string]*google.Subscription
	usubs  map[string]*google.Subscription
}

// NewGoogleClient creates a new connection to Google Pub/Sub.
func NewGoogleClient(ctx context.Context,
	project string, credsPath string) (*GooglePubSub, error) {
	client := &GooglePubSub{
		ctx:    ctx,
		topics: map[string]*google.Topic{},
		subs:   map[string]*google.Subscription{},
		usubs:  map[string]*google.Subscription{},
	}

	// Connect to Pub/Sub.
	var c *google.Client
	var err error
	if credsPath != "emulator" {
		c, err = google.NewClient(ctx, project,
			option.WithCredentialsFile(credsPath))
	} else {
		c, err = google.NewClient(ctx, project)
	}
	if err != nil {
		return nil, err
	}

	client.pubsub = c
	return client, nil
}

// ensureTopic makes sure that a Pub/Sub topic exists.
func (ps *GooglePubSub) ensureTopic(topic string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	t := ps.pubsub.Topic(topic)
	ok, err := t.Exists(ps.ctx)
	if err != nil {
		return err
	}
	if ok {
		ps.topics[topic] = t
		return nil
	}

	// Topic doesn't exist.
	t, err = ps.pubsub.CreateTopic(ps.ctx, topic)
	if err == nil {
		ps.topics[topic] = t
	}
	return err
}

// makeSubscription sets up a Pub/Sub subscription.
func (ps *GooglePubSub) makeSubscription(topic, sub string) (*google.Subscription, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Look up subscription locally.
	subName := topic + "." + sub
	if s, ok := ps.subs[subName]; ok {
		return s, nil
	}

	// Look for subscription on GCP.
	s := ps.pubsub.Subscription(subName)
	ok, err := s.Exists(ps.ctx)
	if err != nil {
		return nil, err
	}
	if ok {
		ps.subs[subName] = s
		return s, nil
	}

	// Subscription doesn't exist, so create it.
	s, err = ps.pubsub.CreateSubscription(ps.ctx, subName,
		google.SubscriptionConfig{Topic: ps.topics[topic]})
	if err != nil {
		return nil, err
	}
	ps.subs[subName] = s
	return s, nil
}

// makeUniqueSubscription sets up a Pub/Sub subscription with a unique
// name for use in fanout.
func (ps *GooglePubSub) makeUniqueSubscription(topic, sub string) (*google.Subscription, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	i := 1
	for {
		subName := fmt.Sprintf("%s.%s-%04x", topic, sub, i)
		i++

		// Look for subscription on GCP.
		s := ps.pubsub.Subscription(subName)
		exists, err := s.Exists(ps.ctx)
		if err != nil {
			return nil, err
		}
		if exists {
			// Name collision: try again...
			continue
		}

		// Create unique subscription.
		s, err = ps.pubsub.CreateSubscription(ps.ctx, subName,
			google.SubscriptionConfig{Topic: ps.topics[topic]})
		if err != nil {
			return nil, err
		}

		ps.usubs[subName] = s
		log.Info().Str("subscription", subName).
			Msg("new pub/sub fanout subscription")

		return s, nil
	}
}

// Publish sends a message on a given topic.
func (ps *GooglePubSub) Publish(topic string, msg interface{}) error {
	err := ps.ensureTopic(topic)
	if err != nil {
		return err
	}
	t, ok := ps.topics[topic]
	if !ok {
		return ErrUnknownTopic
	}

	// Publish message.
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	m := google.Message{Data: data}
	res := t.Publish(ps.ctx, &m)

	// Wait for publication completion and log.
	go func() {
		id, err := res.Get(ps.ctx)
		if err != nil {
			log.Error().Err(err).
				Str("topic", topic).
				Msg("message publication failed")
		} else {
			log.Info().
				Str("topic", topic).
				Str("event-id", id).
				Msg("event published successfully")
		}
	}()

	return nil
}

// Subscribe sets up a subscription, with message content being sent
// on a channel.
func (ps *GooglePubSub) Subscribe(topic, sub string,
	mode SubscriptionMode) (chan []byte, func(), error) {
	err := ps.ensureTopic(topic)
	if err != nil {
		return nil, nil, err
	}
	var subscription *google.Subscription
	if mode == CompetingConsumers {
		subscription, err = ps.makeSubscription(topic, sub)
	} else {
		subscription, err = ps.makeUniqueSubscription(topic, sub)
	}
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithCancel(ps.ctx)
	ch := make(chan []byte)
	go func() {
		err = subscription.Receive(ctx, func(ctx context.Context, m *google.Message) {
			m.Ack()
			ch <- m.Data
		})
		if err != nil {
			log.Error().Err(err).
				Msg("pub/sub Receive returned")
		}
	}()
	return ch, cancel, nil
}

// Close closes off all Pub/Sub topics, sending any unsent messages.
func (ps *GooglePubSub) Close() {
	log.Info().Msg("closing down pub/sub")
	for n, s := range ps.usubs {
		log.Info().Str("subscription", n).Msg("deleting subscription")
		s.Delete(ps.ctx)
	}
	for n, t := range ps.topics {
		log.Info().Str("topic", n).Msg("stopping topic")
		t.Stop()
	}
	ps.topics = map[string]*google.Topic{}
}
