package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrNotFound is returned when a key does not exist.
var ErrNotFound = errors.New("redis: key not found")

// Service is the cache API used by the domain services.
type Service interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Exists(ctx context.Context, key string) (bool, error)
}

type service struct {
	client *redis.Client
}

func New(client *redis.Client) Service {
	return &service{client: client}
}

// Set stores value. Plain strings are stored as is, everything else is
// serialized to JSON.
func (s *service) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	data := value

	if _, ok := value.(string); !ok {
		encoded, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("redis set %q: marshal: %w", key, err)
		}

		data = string(encoded)
	}

	if err := s.client.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("redis set %q: %w", key, err)
	}

	return nil
}

func (s *service) Get(ctx context.Context, key string) (string, error) {
	value, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("redis get %q: %w", key, err)
	}

	return value, nil
}

func (s *service) Del(ctx context.Context, keys ...string) error {
	if err := s.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("redis del %q: %w", keys, err)
	}

	return nil
}

func (s *service) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if err := s.client.Expire(ctx, key, expiration).Err(); err != nil {
		return fmt.Errorf("redis expire %q: %w", key, err)
	}

	return nil
}

func (s *service) Exists(ctx context.Context, key string) (bool, error) {
	count, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists %q: %w", key, err)
	}

	return count > 0, nil
}
