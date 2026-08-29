package session

import "time"

type NewSession struct {
	UserID       int
	RefreshToken string
	UserAgent    string
	IpAddress    string
	ExpiresAt    time.Time
}

type Session struct {
	Id           int
	UserId       int
	CreatedAt    time.Time
	LastUsedAt   time.Time
	ExpiresAt    time.Time
	RefreshToken string
	RevokedAt    time.Time
	UserAgent    string
	IPAddress    string
}
