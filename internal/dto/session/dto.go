package session

import "time"

// NewSession is the input of the session creation flow.
type NewSession struct {
	UserID       int64
	RefreshToken string
	UserAgent    string
	IPAddress    string
	ExpiresAt    time.Time
}

// Session is a stored session.
type Session struct {
	ID           int64
	UserID       int64
	RefreshToken string
	UserAgent    string
	IPAddress    string
	ExpiresAt    time.Time
	CreatedAt    time.Time
	LastUsedAt   *time.Time
	RevokedAt    *time.Time
}
