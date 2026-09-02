package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisService struct {
	redisClient redis.Client
}

type RedisService interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key ...string) error
	Expire(ctx context.Context, key string, duration time.Duration) error
	Exists(ctx context.Context, key string) (bool, error)
}

func NewRedisService(adapter redis.Client) RedisService {
	return &redisService{redisClient: adapter}
}

func (s *redisService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	var data interface{}

	if str, ok := value.(string); ok {
		data = str
	} else {
		jsonData, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("marshal json err: %v", err)
		}
		data = string(jsonData)
	}

	err := s.redisClient.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}

	return nil
}

func (s *redisService) Get(ctx context.Context, key string) (string, error) {
	val, err := s.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("redis get: key not found")
	}
	if err != nil {
		return "", fmt.Errorf("redis get failed: %w", err)
	}

	return val, nil
}

func (s *redisService) Del(ctx context.Context, keys ...string) error {
	err := s.redisClient.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("redis del failed: %w", err)
	}

	return nil
}

func (s *redisService) Exists(ctx context.Context, key string) (bool, error) {
	val, err := s.redisClient.Exists(ctx, key).Result()
	if err == redis.Nil {
		return false, fmt.Errorf("redis exists: key not found")
	}
	if err != nil {
		return false, fmt.Errorf("redis exists failed: %w", err)
	}
	return val > 0, nil
}

func (s *redisService) Expire(ctx context.Context, key string, ttl time.Duration) error {
	err := s.Expire(ctx, key, ttl)
	if err != nil {
		return fmt.Errorf("redis expire failed: %w", err)
	}
	return nil
}
