package auth

import (
	"context"
	authConst "core/internal/const/auth"
	grpcEnum "core/internal/const/grpc"
	authDto "core/internal/dto/auth"
	sessionDto "core/internal/dto/session"
	"core/internal/repository/session"
	"core/internal/service/redis"
	"core/internal/service/token"
	"core/internal/service/user"
	"errors"
	"log/slog"
	"strconv"
	"time"
)

type authService struct {
	log          *slog.Logger
	userService  user.UserSercive
	tokenService token.TokenService
	sessionRepo  session.SessionRepo
	redisService redis.RedisService
}

type AuthService interface {
	Login(ctx context.Context, authType authConst.AuthType, payload authDto.Login) (string, error)
}

func NewAuthService(
	log *slog.Logger,
	userServ user.UserSercive,
	tokenServ *token.TokenService,
	sessionRepo session.SessionRepo,
	redisServ redis.RedisService,
) AuthService {
	return &authService{
		log:          log,
		userService:  userServ,
		tokenService: *tokenServ,
		sessionRepo:  sessionRepo,
		redisService: redisServ,
	}
}

func (s *authService) getUserAgentFromContext(ctx context.Context) string {
	if userAgent, ok := ctx.Value(grpcEnum.UserAgent).(string); ok {
		return userAgent
	}
	return "unknown"
}

func (s *authService) getIPAddressFromContext(ctx context.Context) string {
	if ip, ok := ctx.Value(grpcEnum.IPAddress).(string); ok {
		return ip
	}
	return "unknown"
}

func (s *authService) emailLogin(ctx context.Context, payload authDto.Login) (string, error) {
	userData, err := s.userService.GetUserByEmail(ctx, payload.Email)

	if err != nil {
		return "", errors.New("User not found")
	}

	if !s.userService.CheckPasswordHash(payload.Password, *userData.Password) {
		return "", errors.New("invalid password")
	}

	refreshToken, _, err := s.tokenService.CreateRefreshToken()
	if err != nil {
		return "", err
	}

	sessionData, err := s.sessionRepo.CreateSession(ctx, sessionDto.NewSession{
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(30 * time.Minute),
		UserID:       int(userData.Id),
		UserAgent:    s.getUserAgentFromContext(ctx),
		IpAddress:    s.getIPAddressFromContext(ctx),
	})
	if err != nil {
		return "", err
	}

	accessToken, err := s.tokenService.CreateAccessToken(int(userData.Id), sessionData.Id)

	s.saveSessionRedisAsync(sessionData)

	return accessToken, nil
}

func (s *authService) saveSessionRedisAsync(sessionData *sessionDto.Session) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		sessionKey := "session-" + strconv.Itoa(sessionData.UserId)

		if err := s.redisService.Set(ctx, sessionKey, strconv.Itoa(sessionData.UserId), 20*time.Minute); err != nil {
			s.log.Error("Redis save failed",
				slog.String("sessionKey", sessionKey),
				slog.String("error", err.Error()),
			)
		}
	}()
}

func (s *authService) Login(ctx context.Context, authType authConst.AuthType, payload authDto.Login) (string, error) {
	switch authType {
	case authConst.EmailAuth:
		tokenStr, err := s.emailLogin(ctx, authDto.Login{Email: payload.Email, Password: payload.Password})
		if err != nil {
			return "", err
		}
		return tokenStr, nil
	}
	return "", errors.New("invalid auth type")
}
