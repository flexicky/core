package token

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	SessionId string `json:"session_id"`
	jwt.RegisteredClaims
}

type TokenService struct {
	privateKey ed25519.PrivateKey
}

func NewTokenService(privateKey ed25519.PrivateKey) *TokenService {
	return &TokenService{privateKey: privateKey}
}

func (s *TokenService) CreateAccessToken(userId, sessionId int) (string, error) {
	now := time.Now()

	claims := &Claims{
		SessionId: strconv.Itoa(sessionId),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(userId),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(20 * time.Minute)),
			ID:        strconv.Itoa(userId),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(s.privateKey)
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
