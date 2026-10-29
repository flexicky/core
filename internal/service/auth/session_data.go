package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"core/internal/apperr"

	sessionDto "core/internal/dto/session"
	userDto "core/internal/dto/user"
)

// ErrCacheMiss is returned when the session is not present in the cache.
var ErrCacheMiss = errors.New("auth: session cache miss")

// SessionData is the session representation cached in Redis and returned to
// the caller of the "me" endpoint.
type SessionData struct {
	UserID           int64   `json:"user_id"`
	UserName         *string `json:"user_name"`
	UserEmail        *string `json:"user_email"`
	SessionID        int64   `json:"session_id"`
	ExpiresAt        int64   `json:"expires_at"`
	UserCreatedAt    int64   `json:"user_created_at"`
	SessionCreatedAt int64   `json:"session_created_at"`
}

// SessionDataFrom builds the cache payload out of a user and its session.
func SessionDataFrom(user *userDto.User, session *sessionDto.Session) SessionData {
	data := SessionData{
		UserID:           user.ID,
		UserName:         user.Name,
		UserEmail:        user.Email,
		SessionID:        session.ID,
		ExpiresAt:        session.ExpiresAt.Unix(),
		UserCreatedAt:    user.CreatedAt.Unix(),
		SessionCreatedAt: session.CreatedAt.Unix(),
	}

	return data
}

func parseSessionID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid session id %q", apperr.ErrUnauthorized, raw)
	}

	return id, nil
}

func unmarshalSession(raw string) (*SessionData, error) {
	var data SessionData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, fmt.Errorf("unmarshal cached session: %w", err)
	}

	return &data, nil
}

// sessionKey builds the cache key. The "session-" prefix is part of the
// established key format, do not change it without migrating the data.
func sessionKey(prefix, sessionID string) string {
	return prefix + "session-" + sessionID
}

// formatSessionID renders the numeric session id the way it is used in tokens
// and cache keys.
func formatSessionID(id int64) string {
	return strconv.FormatInt(id, 10)
}
