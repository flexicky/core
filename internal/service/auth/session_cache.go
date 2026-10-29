package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	redisServ "core/internal/service/redis"
)

// SessionCache stores session snapshots so that the "me" endpoint rarely
// touches the database.
type SessionCache interface {
	// Save persists data without blocking the caller.
	Save(ctx context.Context, data SessionData)
	// Get returns the cached session or ErrCacheMiss.
	Get(ctx context.Context, sessionID string) (*SessionData, error)
}

type cache struct {
	log    *slog.Logger
	redis  redisServ.Service
	prefix string
	ttl    time.Duration
}

// NewSessionCache builds the Redis backed session cache.
func NewSessionCache(log *slog.Logger, redis redisServ.Service, prefix string, ttl time.Duration) SessionCache {
	return &cache{log: log, redis: redis, prefix: prefix, ttl: ttl}
}

// Save writes the session in the background: a cache outage must never fail a
// login, but it must be visible in the logs.
func (c *cache) Save(ctx context.Context, data SessionData) {
	// Detach from the request context: the write outlives the gRPC call.
	ctx = context.WithoutCancel(ctx)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.log.Error("panic while caching session", slog.Any("recover", r), slog.Int64("session_id", data.SessionID))
			}
		}()

		key := sessionKey(c.prefix, formatSessionID(data.SessionID))

		if err := c.redis.Set(ctx, key, data, c.ttl); err != nil {
			c.log.Error("cache session", slog.String("key", key), slog.Any("error", err))
		}
	}()
}

func (c *cache) Get(ctx context.Context, sessionID string) (*SessionData, error) {
	raw, err := c.redis.Get(ctx, sessionKey(c.prefix, sessionID))
	if errors.Is(err, redisServ.ErrNotFound) {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}

	return unmarshalSession(raw)
}
