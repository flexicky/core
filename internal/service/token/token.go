package token

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"

	"core/internal/apperr"
	tokenDto "core/internal/dto/token"

	"github.com/golang-jwt/jwt/v5"
)

// SeedSize is the size of an ed25519 seed in bytes.
const SeedSize = ed25519.SeedSize

// Service issues and validates access tokens.
type Service interface {
	CreateAccessToken(userID, sessionID int64) (token string, expiresAt time.Time, err error)
	CreateRefreshToken() (token string, hash string, err error)
	ParseAccessToken(token string) (*tokenDto.Claims, error)
}

type service struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	ttl        time.Duration
	now        func() time.Time
}

func New(privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey, ttl time.Duration) Service {
	return &service{
		privateKey: privateKey,
		publicKey:  publicKey,
		ttl:        ttl,
		now:        time.Now,
	}
}

// KeysFromSeed derives a stable ed25519 key pair from a 32 byte seed.
func KeysFromSeed(seed []byte) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	if len(seed) != SeedSize {
		return nil, nil, fmt.Errorf("ed25519 seed must be %d bytes, got %d", SeedSize, len(seed))
	}

	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey, ok := privateKey.Public().(ed25519.PublicKey)
	if !ok {
		return nil, nil, errors.New("unexpected public key type")
	}

	return privateKey, publicKey, nil
}

// GenerateKeys creates a new random ed25519 key pair.
func GenerateKeys() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate ed25519 keys: %w", err)
	}

	return privateKey, publicKey, nil
}

func (s *service) CreateAccessToken(userID, sessionID int64) (string, time.Time, error) {
	now := s.now()
	expiresAt := now.Add(s.ttl)

	claims := &tokenDto.Claims{
		SessionID: strconv.FormatInt(sessionID, 10),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			ID:        strconv.FormatInt(userID, 10),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(s.privateKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}

	return signed, expiresAt, nil
}

// CreateRefreshToken returns an opaque refresh token and the hash that should
// be persisted instead of the token itself.
func (s *service) CreateRefreshToken() (string, string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	refreshToken := base64.RawURLEncoding.EncodeToString(buf)
	hash := sha256.Sum256([]byte(refreshToken))

	return refreshToken, base64.RawURLEncoding.EncodeToString(hash[:]), nil
}

func (s *service) ParseAccessToken(raw string) (*tokenDto.Claims, error) {
	claims := &tokenDto.Claims{}

	parsed, err := jwt.ParseWithClaims(
		raw,
		claims,
		func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodEdDSA {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Method.Alg())
			}

			return s.publicKey, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
	)
	if err != nil {
		return nil, errors.Join(apperr.ErrInvalidToken, err)
	}

	if !parsed.Valid {
		return nil, apperr.ErrInvalidToken
	}

	return claims, nil
}
