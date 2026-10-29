package token

import (
	"errors"
	"testing"
	"time"

	"core/internal/apperr"
)

func newTestService(t *testing.T, ttl time.Duration) Service {
	t.Helper()

	privateKey, publicKey, err := GenerateKeys()
	if err != nil {
		t.Fatalf("GenerateKeys() error = %v", err)
	}

	return New(privateKey, publicKey, ttl)
}

func TestAccessTokenRoundTrip(t *testing.T) {
	t.Parallel()

	svc := newTestService(t, time.Hour)

	raw, expiresAt, err := svc.CreateAccessToken(42, 7)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}

	if raw == "" {
		t.Fatal("CreateAccessToken() returned an empty token")
	}

	if time.Until(expiresAt) < 59*time.Minute {
		t.Fatalf("expiresAt = %v, want about one hour from now", expiresAt)
	}

	claims, err := svc.ParseAccessToken(raw)
	if err != nil {
		t.Fatalf("ParseAccessToken() error = %v", err)
	}

	if claims.SessionID != "7" {
		t.Fatalf("SessionID = %q, want %q", claims.SessionID, "7")
	}

	if claims.Subject != "42" {
		t.Fatalf("Subject = %q, want %q", claims.Subject, "42")
	}
}

func TestParseAccessTokenRejectsGarbage(t *testing.T) {
	t.Parallel()

	svc := newTestService(t, time.Hour)

	_, err := svc.ParseAccessToken("not-a-token")
	if !errors.Is(err, apperr.ErrInvalidToken) {
		t.Fatalf("ParseAccessToken() error = %v, want ErrInvalidToken", err)
	}
}

func TestParseAccessTokenRejectsWrongKey(t *testing.T) {
	t.Parallel()

	issuer := newTestService(t, time.Hour)
	verifier := newTestService(t, time.Hour)

	raw, _, err := issuer.CreateAccessToken(1, 1)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}

	if _, err := verifier.ParseAccessToken(raw); !errors.Is(err, apperr.ErrInvalidToken) {
		t.Fatalf("ParseAccessToken() error = %v, want ErrInvalidToken", err)
	}
}

func TestParseAccessTokenRejectsExpiredToken(t *testing.T) {
	t.Parallel()

	privateKey, publicKey, err := GenerateKeys()
	if err != nil {
		t.Fatalf("GenerateKeys() error = %v", err)
	}

	svc := New(privateKey, publicKey, time.Hour).(*service)
	svc.now = func() time.Time { return time.Now().Add(-2 * time.Hour) }

	raw, _, err := svc.CreateAccessToken(1, 1)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}

	if _, err := svc.ParseAccessToken(raw); !errors.Is(err, apperr.ErrInvalidToken) {
		t.Fatalf("ParseAccessToken() error = %v, want ErrInvalidToken", err)
	}
}

func TestCreateRefreshTokenIsRandom(t *testing.T) {
	t.Parallel()

	svc := newTestService(t, time.Hour)

	first, firstHash, err := svc.CreateRefreshToken()
	if err != nil {
		t.Fatalf("CreateRefreshToken() error = %v", err)
	}

	second, secondHash, err := svc.CreateRefreshToken()
	if err != nil {
		t.Fatalf("CreateRefreshToken() error = %v", err)
	}

	if first == second || firstHash == secondHash {
		t.Fatal("refresh tokens must not repeat")
	}

	if firstHash == first {
		t.Fatal("the stored hash must differ from the refresh token itself")
	}
}

func TestKeysFromSeedIsDeterministic(t *testing.T) {
	t.Parallel()

	seed := make([]byte, SeedSize)
	for i := range seed {
		seed[i] = byte(i)
	}

	firstPrivate, firstPublic, err := KeysFromSeed(seed)
	if err != nil {
		t.Fatalf("KeysFromSeed() error = %v", err)
	}

	secondPrivate, secondPublic, err := KeysFromSeed(seed)
	if err != nil {
		t.Fatalf("KeysFromSeed() error = %v", err)
	}

	if !firstPrivate.Equal(secondPrivate) || !firstPublic.Equal(secondPublic) {
		t.Fatal("KeysFromSeed() must be deterministic")
	}

	if _, _, err := KeysFromSeed([]byte("short")); err == nil {
		t.Fatal("KeysFromSeed() must reject seeds of the wrong length")
	}
}
