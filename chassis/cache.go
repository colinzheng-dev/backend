package chassis

import (
	"encoding/json"

	lru "github.com/hashicorp/golang-lru"
	"github.com/rs/zerolog/log"

	"github.com/veganbase/backend/chassis/pubsub"
)

// Cache represents an LRU cache keyed by unique strings that supports
// invalidation messages sent over a pub/sub topic.
type Cache struct {
	lru          *lru.Cache
	pubsub       pubsub.PubSub
	invalTopic   string
	invalSubName string
}

// NewCache creates a new cache of the given size, evicting items when
// messages are received on the given pub/sub topic (invaldation
// messages contain a single key to invalidate, as a string).
func NewCache(size int, ps pubsub.PubSub, invalTopic string, appName string) (*Cache, error) {
	// Create new cache.
	cache := Cache{
		pubsub:     ps,
		invalTopic: invalTopic,
	}
	lru, err := lru.New(size)
	if err != nil {
		return nil, err
	}
	cache.lru = lru

	// Set up invalidator goroutine.
	go cache.invalidator(invalTopic, appName)

	return &cache, nil
}

// Get gets a value from the cache, returning the value and a flag to
// say whether the item was present.
func (cache *Cache) Get(key string) (interface{}, bool) {
	return cache.lru.Get(key)
}

// Set sets a value in the cache.
func (cache *Cache) Set(key string, val interface{}) {
	cache.lru.Add(key, val)
}

// Delete deletes a key from the cache.
func (cache *Cache) Delete(key string) {
	cache.lru.Remove(key)
}

// Invalidate cache items in response to pub/sub messages.
func (cache *Cache) invalidator(invalTopic, appName string) {
	invalSubName := invalTopic + "-" + appName
	invalCh, _, err := cache.pubsub.Subscribe(invalTopic, invalSubName, pubsub.Fanout)
	if err != nil {
		log.Fatal().Err(err).
			Msg("unable to subscribe to cache invalidation topic '" + invalTopic + "'")
	}

	for {
		idJSON := <-invalCh
		var id string
		err := json.Unmarshal(idJSON, &id)
		if err != nil {
			log.Error().Err(err).
				Msg("decoding cache invalidation message")
			continue
		}
		cache.Delete(string(id))
	}
}
