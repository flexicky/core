package token

import (
	tokenDto "core/internal/dto/token"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

func NewTokenService(privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) *TokenService {
	return &TokenService{privateKey: privateKey, publicKey: publicKey}
}

func (s *TokenService) CreateAccessToken(userId, sessionId int) (string, error) {
	now := time.Now()

	claims := &tokenDto.Claims{
		SessionId: strconv.Itoa(sessionId),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(userId),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(20 * time.Minute)),
			ID:        strconv.Itoa(userId),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)

	signedToken, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (s *TokenService) CreateRefreshToken() (token string, hash string, err error) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	token = base64.RawURLEncoding.EncodeToString(bytes)

	hashBytes := sha256.Sum256([]byte(token))

	hash = base64.RawURLEncoding.EncodeToString(hashBytes[:])

	return token, hash, nil
}

func (s *TokenService) ParseAccessToken(token string) (*tokenDto.Claims, error) {
	claims := &tokenDto.Claims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodEdDSA {
			return nil, errors.New("unexpected signing method")
		}
		return s.publicKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !tkn.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
