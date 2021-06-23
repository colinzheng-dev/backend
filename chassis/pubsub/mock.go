package pubsub

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
)

type MockPubSub struct {
	M map[string][][]byte
}

func NewMockPubSub(messages map[string][][]byte) *MockPubSub {
	return &MockPubSub{messages}
}

func (mock *MockPubSub) Publish(topic string, message interface{}) error {
	fmt.Println("MOCK Publish", topic, "   ", message)
	msg, err := json.Marshal(message)
	if err != nil {
		return err
	}
	log.Info().
		Str("topic", topic).
		Interface("message", message).
		Msg("message published")
	mock.M[topic] = append(mock.M[topic], msg)
	return nil
}

func (mock *MockPubSub) Subscribe(topic, sub string,
	mode SubscriptionMode) (chan []byte, func(), error) {
	return nil, nil, nil
}

func (mock *MockPubSub) Close() {
}
