package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"core/internal/apperr"
	authConst "core/internal/const/auth"
	"core/internal/ctxkeys"
	authDto "core/internal/dto/auth"
	sessionDto "core/internal/dto/session"
	sessionServ "core/internal/service/session"
	"core/internal/service/token"
	userServ "core/internal/service/user"
)

// Service is the authentication domain API.
type Service interface {
	// Login verifies the credentials and issues an access token.
	Login(ctx context.Context, authType authConst.AuthType, payload authDto.Login) (*authDto.LoginResult, error)
	// CurrentSession returns the session referenced by the access token
	// carried in ctx.
	CurrentSession(ctx context.Context) (*SessionData, error)
}

type service struct {
	log        *slog.Logger
	users      userServ.Service
	tokens     token.Service
	sessions   sessionServ.Service
	cache      SessionCache
	sessionTTL time.Duration
}

func New(
	log *slog.Logger,
	users userServ.Service,
	tokens token.Service,
	sessions sessionServ.Service,
	cache SessionCache,
	sessionTTL time.Duration,
) Service {
	return &service{
		log:        log,
		users:      users,
		tokens:     tokens,
		sessions:   sessions,
		cache:      cache,
		sessionTTL: sessionTTL,
	}
}

func (s *service) Login(ctx context.Context, authType authConst.AuthType, payload authDto.Login) (*authDto.LoginResult, error) {
	switch authType {
	case authConst.EmailAuth:
		return s.emailLogin(ctx, payload)
	default:
		return nil, apperr.ErrInvalidAuthType
	}
}

func (s *service) emailLogin(ctx context.Context, payload authDto.Login) (*authDto.LoginResult, error) {
	const op = "auth.emailLogin"

	user, err := s.users.GetByEmail(ctx, payload.Email)
	if err != nil {
		if errors.Is(err, apperr.ErrUserNotFound) {
			return nil, apperr.ErrInvalidCredentials
		}

		return nil, fmt.Errorf("%s: get user: %w", op, err)
	}

	if user.Password == nil || !s.users.CheckPassword(payload.Password, *user.Password) {
		return nil, apperr.ErrInvalidCredentials
	}

	refreshToken, _, err := s.tokens.CreateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	client := ctxkeys.ClientInfoFrom(ctx)

	session, err := s.sessions.Create(ctx, sessionDto.NewSession{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    client.UserAgent,
		IPAddress:    client.IP,
		ExpiresAt:    time.Now().Add(s.sessionTTL),
	})
	if err != nil {
		return nil, fmt.Errorf("%s: create session: %w", op, err)
	}

	accessToken, expiresAt, err := s.tokens.CreateAccessToken(user.ID, session.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: create access token: %w", op, err)
	}

	s.cache.Save(ctx, SessionDataFrom(user, session))

	return &authDto.LoginResult{
		AccessToken: accessToken,
		ExpiresAt:   expiresAt,
	}, nil
}

func (s *service) CurrentSession(ctx context.Context) (*SessionData, error) {
	const op = "auth.CurrentSession"

	sessionID, ok := ctxkeys.SessionID(ctx)
	if !ok {
		return nil, apperr.ErrUnauthorized
	}

	cached, err := s.cache.Get(ctx, sessionID)
	if err == nil {
		return cached, nil
	}

	if !errors.Is(err, ErrCacheMiss) {
		s.log.WarnContext(ctx, "session cache unavailable, falling back to database", slog.String("op", op), slog.Any("error", err))
	}

	data, err := s.loadSessionFromDB(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	s.cache.Save(ctx, *data)

	return data, nil
}

func (s *service) loadSessionFromDB(ctx context.Context, sessionID string) (*SessionData, error) {
	const op = "auth.loadSessionFromDB"

	id, err := parseSessionID(sessionID)
	if err != nil {
		return nil, err
	}

	session, err := s.sessions.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user, err := s.users.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	data := SessionDataFrom(user, session)

	return &data, nil
}
