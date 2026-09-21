package auth

import (
	"context"
	authConst "core/internal/const/auth"
	grpcEnum "core/internal/const/grpc"
	authDto "core/internal/dto/auth"
	sessionDto "core/internal/dto/session"
	"core/internal/service/redis"
	sessionServ "core/internal/service/session"
	"core/internal/service/token"
	"core/internal/service/user"
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"time"
)

type authService struct {
	log            *slog.Logger
	userService    user.UserSercive
	tokenService   token.TokenService
	sessionService sessionServ.SessionService
	redisService   redis.RedisService
}

type AuthService interface {
	Login(ctx context.Context, authType authConst.AuthType, payload authDto.Login) (*authDto.LoginResult, error)
	GetCurrentSession(ctx context.Context) (user *SessionJsonData, err error)
}

func NewAuthService(
	log *slog.Logger,
	userServ user.UserSercive,
	tokenServ *token.TokenService,
	sessionService sessionServ.SessionService,
	redisServ redis.RedisService,
) AuthService {
	return &authService{
		log:            log,
		userService:    userServ,
		tokenService:   *tokenServ,
		sessionService: sessionService,
		redisService:   redisServ,
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

func (s *authService) emailLogin(ctx context.Context, payload authDto.Login) (*authDto.LoginResult, error) {
	userData, err := s.userService.GetUserByEmail(ctx, payload.Email)

	if err != nil {
		return nil, errors.New("User not found")
	}

	if !s.userService.CheckPasswordHash(payload.Password, *userData.Password) {
		return nil, errors.New("invalid password")
	}

	refreshToken, _, err := s.tokenService.CreateRefreshToken()
	if err != nil {
		return nil, err
	}

	sessionData, err := s.sessionService.CreateSession(ctx, sessionDto.NewSession{
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(30 * time.Minute),
		UserID:       int(userData.Id),
		UserAgent:    s.getUserAgentFromContext(ctx),
		IpAddress:    s.getIPAddressFromContext(ctx),
	})
	if err != nil {
		return nil, err
	}

	accessToken, err := s.tokenService.CreateAccessToken(int(userData.Id), sessionData.Id)
	if err != nil {
		return nil, err
	}

	s.saveSessionRedisAsync(sessionData)

	result := &authDto.LoginResult{
		Access_token: accessToken,
		Expires_At:   time.Now().Add(20 * time.Minute),
	}

	return result, nil
}

type SessionJsonData struct {
	UserId           int     `json:"user_id"`
	UserName         *string `json:"user_name"`
	UserEmail        *string `json:"user_email"`
	SessionId        int     `json:"session_id"`
	ExpiresAt        int64   `json:"expires_at"`
	UserCreatedAt    int64   `json:"user_created_at"`
	SessionCreatedAt int64   `json:"session_created_at"`
}

func (s *authService) saveSessionRedisAsync(sessionData *sessionDto.Session) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.log.Error("panic in saveSessionRedisAsync", "recover", r)
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		userData, err := s.userService.GetUserById(ctx, int(sessionData.UserId))
		if err != nil {
			s.log.Warn("Failed to get user by id", "userId", sessionData.UserId, "err", err)
			return
		}

		redisData := map[string]interface{}{
			"user_id":            userData.Id,
			"session_id":         sessionData.Id,
			"expires_at":         sessionData.ExpiresAt,
			"session_created":    sessionData.CreatedAt,
			"user_name":          userData.Name,
			"user_email":         userData.Email,
			"user_created_at":    userData.CreatedAt,
			"session_created_at": sessionData.CreatedAt,
		}

		jsonData, err := json.Marshal(redisData)
		if err != nil {
			s.log.Error("Failed to marshal session data",
				slog.String("sessionId", strconv.Itoa(sessionData.Id)),
				slog.String("error", err.Error()))
			return
		}

		sessionKey := "session-" + strconv.Itoa(sessionData.Id)

		if err := s.redisService.Set(ctx, sessionKey, string(jsonData), 20*time.Minute); err != nil {
			s.log.Error("Redis save failed",
				slog.String("sessionKey", sessionKey),
				slog.String("error", err.Error()),
			)
		}
	}()
}

func (s *authService) GetCurrentSession(ctx context.Context) (user *SessionJsonData, err error) {
	sessionId, ok := ctx.Value("session_id").(string)
	if !ok {
		return nil, errors.New("sessionId not found")
	}
	sessionKey := "session-" + sessionId
	sessionData, err := s.redisService.Get(ctx, sessionKey)
	if err != nil {
		intSessionId, err := strconv.Atoi(sessionId)
		if err != nil {
			return nil, err
		}
		currentSession, err := s.sessionService.GetSessionById(ctx, intSessionId)
		if err != nil {
			return nil, err
		}
		currentUser, err := s.userService.GetUserById(ctx, currentSession.UserId)
		if err != nil {
			return nil, err
		}

		s.saveSessionRedisAsync(currentSession)

		return &SessionJsonData{
			UserId:           currentSession.UserId,
			UserName:         currentUser.Name,
			UserEmail:        currentUser.Email,
			SessionId:        intSessionId,
			ExpiresAt:        currentSession.ExpiresAt.Unix(),
			SessionCreatedAt: currentSession.CreatedAt.Unix(),
			UserCreatedAt:    currentUser.CreatedAt.Unix(),
		}, nil
	}

	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(jsonData, &user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, authType authConst.AuthType, payload authDto.Login) (*authDto.LoginResult, error) {
	switch authType {
	case authConst.EmailAuth:
		tokenStr, err := s.emailLogin(ctx, authDto.Login{Email: payload.Email, Password: payload.Password})
		if err != nil {
			return nil, err
		}
		return tokenStr, nil
	}
	return nil, errors.New("invalid auth type")
}
