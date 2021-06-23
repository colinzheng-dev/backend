package chassis

import (
	"github.com/go-redis/redis"
	"time"
)

// locker implementation with Redis
type RedisClient struct {
	cache *redis.Client
}

func NewRedisClient(address, pwd string) RedisClient {
	return RedisClient{
		cache: redis.NewClient(&redis.Options{
			Addr:     address,
			Password: pwd,
		}),
	}
}

func (s *RedisClient) Lock(key string) (bool, error) {
	res, err := s.cache.SetNX(key, time.Now().String(), time.Second*15).Result()
	if err != nil {
		return false, err
	}
	return res, nil
}

func (s *RedisClient) Unlock(key string) error {
	return s.cache.Del(key).Err()
}
