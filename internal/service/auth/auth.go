package auth

import (
	"context"
	authConst "core/internal/const/auth"
	authDto "core/internal/dto/auth"
	sessionDto "core/internal/dto/session"
	"core/internal/repository/session"
	"core/internal/service/token"
	"core/internal/service/user"
	"errors"
	"fmt"
	"time"
)

type authService struct {
	userService  user.UserSercive
	tokenService token.TokenService
	sessionRepo  session.SessionRepo
}

type AuthService interface {
	Login(ctx context.Context, authType authConst.AuthType, payload authDto.Login) (string, error)
}

func NewAuthService(
	userServ user.UserSercive,
	tokenServ *token.TokenService,
	sessionRepo session.SessionRepo,
) AuthService {
	return &authService{
		userService:  userServ,
		tokenService: *tokenServ,
		sessionRepo:  sessionRepo,
	}
}

func (s *authService) emailLogin(ctx context.Context, payload authDto.Login) (string, error) {
	userData, err := s.userService.GetUserByEmail(ctx, payload.Email)
	fmt.Println(userData)
	if err != nil {
		return "", err
	}

	if !s.userService.CheckPasswordHash(payload.Password, *userData.Password) {
		return "", errors.New("invalid password")
	}

	if *userData.Email != payload.Email {
		return "", errors.New("email does not match")
	}

	refreshToken, _, err := s.tokenService.CreateRefreshToken()
	if err != nil {
		return "", err
	}

	sessionData, err := s.sessionRepo.CreateSession(ctx, sessionDto.NewSession{
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(30 * time.Minute),
		UserID:       int(userData.Id),
		UserAgent:    "someAgent",
		IpAddress:    "127.0.0.1",
	})
	if err != nil {
		return "", err
	}

	accessToken, err := s.tokenService.CreateAccessToken(int(userData.Id), sessionData.Id)

	return accessToken, nil
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
