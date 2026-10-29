// Package session_cleaner removes expired sessions from the database.
package session_cleaner

import (
	"context"
	"log/slog"
	"time"

	sessionServ "core/internal/service/session"
)

// Process periodically deletes expired sessions.
type Process struct {
	log            *slog.Logger
	sessionService sessionServ.Service
	interval       time.Duration
}

func NewProcess(log *slog.Logger, sessionService sessionServ.Service, interval time.Duration) *Process {
	return &Process{
		log:            log,
		sessionService: sessionService,
		interval:       interval,
	}
}

// Run blocks until ctx is cancelled, cleaning up once immediately and then
// every interval.
func (p *Process) Run(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	p.cleanup(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.cleanup(ctx)
		}
	}
}

func (p *Process) cleanup(ctx context.Context) {
	deleted, err := p.sessionService.DeleteExpired(ctx, time.Now())
	if err != nil {
		p.log.ErrorContext(ctx, "delete expired sessions", slog.Any("error", err))

		return
	}

	if deleted > 0 {
		p.log.InfoContext(ctx, "expired sessions deleted", slog.Int64("count", deleted))
	}
}
