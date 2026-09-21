package session_cleaner

import (
	"context"
	sessionServ "core/internal/service/session"
	"core/internal/utils"
	"log/slog"
	"time"
)

type Process struct {
	log            *slog.Logger
	sessionService sessionServ.SessionService
	MaxRetries     int
	TimeSleep      time.Duration
}

func NewProcess(log *slog.Logger, maxRetr int, timeSleep time.Duration, sessionService sessionServ.SessionService) *Process {
	return &Process{
		log:            log,
		MaxRetries:     maxRetr,
		TimeSleep:      timeSleep,
		sessionService: sessionService,
	}
}

func (p *Process) cleanSession(ctx context.Context) {

	deadSessions, err := p.sessionService.GetSessionByExpiresA(ctx, time.Now())

	if err != nil {
		p.log.Error("GetSessionByExpiresA", "err", err)
		return
	}

	if len(deadSessions) > 0 {
		for _, deadSession := range deadSessions {
			if err := p.sessionService.DeleteSessionById(ctx, int(deadSession)); err != nil {
				p.log.Error("Error DeleteSessionById", "err", err)
				continue
			}
			p.log.Info("DeleteSessionById", "deadSession", deadSession)
		}
	}
}

func (p *Process) Run(ctx context.Context) {
	utils.Process(ctx, p.TimeSleep, p.MaxRetries, func(ctx context.Context) {
		p.cleanSession(ctx)
	})
}
