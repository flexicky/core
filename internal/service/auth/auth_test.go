package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"core/internal/apperr"
	authConst "core/internal/const/auth"
	"core/internal/ctxkeys"
	authDto "core/internal/dto/auth"
	sessionDto "core/internal/dto/session"
	userDto "core/internal/dto/user"
	redisServ "core/internal/service/redis"
	"core/internal/service/token"
)

// --- fakes -----------------------------------------------------------------

type fakeUsers struct {
	byEmail map[string]*userDto.User
	byID    map[int64]*userDto.User
	err     error
}

func (f *fakeUsers) Create(context.Context, userDto.NewUser) (*userDto.User, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeUsers) GetByID(_ context.Context, id int64) (*userDto.User, error) {
	if f.err != nil {
		return nil, f.err
	}

	user, ok := f.byID[id]
	if !ok {
		return nil, apperr.ErrUserNotFound
	}

	return user, nil
}

func (f *fakeUsers) GetByEmail(_ context.Context, email string) (*userDto.User, error) {
	if f.err != nil {
		return nil, f.err
	}

	user, ok := f.byEmail[email]
	if !ok {
		return nil, apperr.ErrUserNotFound
	}

	return user, nil
}

func (f *fakeUsers) CheckPassword(password, hash string) bool {
	return hash == "hash:"+password
}

type fakeSessions struct {
	created  []sessionDto.NewSession
	sessions map[int64]*sessionDto.Session
	err      error
}

func (f *fakeSessions) Create(_ context.Context, params sessionDto.NewSession) (*sessionDto.Session, error) {
	if f.err != nil {
		return nil, f.err
	}

	f.created = append(f.created, params)

	created := &sessionDto.Session{
		ID:           int64(len(f.created)),
		UserID:       params.UserID,
		RefreshToken: params.RefreshToken,
		UserAgent:    params.UserAgent,
		IPAddress:    params.IPAddress,
		ExpiresAt:    params.ExpiresAt,
		CreatedAt:    time.Now(),
	}

	if f.sessions == nil {
		f.sessions = make(map[int64]*sessionDto.Session)
	}

	f.sessions[created.ID] = created

	return created, nil
}

func (f *fakeSessions) GetByID(_ context.Context, id int64) (*sessionDto.Session, error) {
	if f.err != nil {
		return nil, f.err
	}

	found, ok := f.sessions[id]
	if !ok {
		return nil, apperr.ErrSessionNotFound
	}

	return found, nil
}

func (f *fakeSessions) DeleteExpired(context.Context, time.Time) (int64, error) {
	return 0, nil
}

type fakeRedis struct {
	values map[string]string
	err    error
}

func (f *fakeRedis) Set(_ context.Context, key string, value any, _ time.Duration) error {
	if f.err != nil {
		return f.err
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if f.values == nil {
		f.values = make(map[string]string)
	}

	f.values[key] = string(encoded)

	return nil
}

func (f *fakeRedis) Get(_ context.Context, key string) (string, error) {
	if f.err != nil {
		return "", f.err
	}

	value, ok := f.values[key]
	if !ok {
		return "", redisServ.ErrNotFound
	}

	return value, nil
}

func (f *fakeRedis) Del(context.Context, ...string) error { return nil }

func (f *fakeRedis) Expire(context.Context, string, time.Duration) error { return nil }

func (f *fakeRedis) Exists(context.Context, string) (bool, error) { return false, nil }

// --- helpers ---------------------------------------------------------------

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestTokens(t *testing.T) token.Service {
	t.Helper()

	privateKey, publicKey, err := token.GenerateKeys()
	if err != nil {
		t.Fatalf("GenerateKeys() error = %v", err)
	}

	return token.New(privateKey, publicKey, time.Hour)
}

type fixture struct {
	service  Service
	users    *fakeUsers
	sessions *fakeSessions
	redis    *fakeRedis
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	email := "user@example.com"
	name := "Test User"
	hash := "hash:secret123"

	users := &fakeUsers{
		byEmail: map[string]*userDto.User{
			email: {ID: 1, Name: &name, Email: &email, Password: &hash, CreatedAt: time.Now()},
		},
		byID: map[int64]*userDto.User{
			1: {ID: 1, Name: &name, Email: &email, CreatedAt: time.Now()},
		},
	}

	sessions := &fakeSessions{}
	redis := &fakeRedis{}

	cache := NewSessionCache(testLogger(), redis, "test", time.Minute)

	return &fixture{
		service:  New(testLogger(), users, newTestTokens(t), sessions, cache, time.Hour),
		users:    users,
		sessions: sessions,
		redis:    redis,
	}
}

// --- tests -----------------------------------------------------------------

func TestLoginIssuesTokenAndSession(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	ctx := ctxkeys.WithClientInfo(context.Background(), ctxkeys.ClientInfo{
		UserAgent: "grpc-go/1.83",
		IP:        "10.0.0.1",
	})

	result, err := f.service.Login(ctx, authConst.EmailAuth, authDto.Login{
		Email:    "user@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if result.AccessToken == "" {
		t.Fatal("Login() returned an empty access token")
	}

	if time.Until(result.ExpiresAt) < 30*time.Minute {
		t.Fatalf("ExpiresAt = %v, want about one hour from now", result.ExpiresAt)
	}

	if len(f.sessions.created) != 1 {
		t.Fatalf("sessions created = %d, want 1", len(f.sessions.created))
	}

	created := f.sessions.created[0]
	if created.UserID != 1 {
		t.Fatalf("UserID = %d, want 1", created.UserID)
	}

	if created.UserAgent != "grpc-go/1.83" || created.IPAddress != "10.0.0.1" {
		t.Fatalf("client info not propagated: %+v", created)
	}

	if created.RefreshToken == "" {
		t.Fatal("session must carry a refresh token")
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	_, err := f.service.Login(context.Background(), authConst.EmailAuth, authDto.Login{
		Email:    "user@example.com",
		Password: "wrong-password",
	})
	if !errors.Is(err, apperr.ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}

	if len(f.sessions.created) != 0 {
		t.Fatal("no session may be created for a failed login")
	}
}

func TestLoginHidesUnknownUser(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	_, err := f.service.Login(context.Background(), authConst.EmailAuth, authDto.Login{
		Email:    "missing@example.com",
		Password: "secret123",
	})
	if !errors.Is(err, apperr.ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestLoginRejectsUnsupportedAuthType(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	_, err := f.service.Login(context.Background(), authConst.TelegramAuth, authDto.Login{})
	if !errors.Is(err, apperr.ErrInvalidAuthType) {
		t.Fatalf("Login() error = %v, want ErrInvalidAuthType", err)
	}
}

func TestLoginSucceedsForUserWithoutPassword(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	email := "user@example.com"
	f.users.byEmail[email] = &userDto.User{ID: 1, Email: &email, CreatedAt: time.Now()}

	// A user without a password must not panic and must not authenticate.
	if _, err := f.service.Login(context.Background(), authConst.EmailAuth, authDto.Login{
		Email:    email,
		Password: "secret123",
	}); !errors.Is(err, apperr.ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestCurrentSessionUsesCache(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	name := "Cached User"
	email := "cached@example.com"
	f.redis.values = map[string]string{
		sessionKey("test", "5"): `{"user_id":9,"user_name":"Cached User","user_email":"cached@example.com","session_id":5}`,
	}

	ctx := ctxkeys.WithSessionID(context.Background(), "5")

	data, err := f.service.CurrentSession(ctx)
	if err != nil {
		t.Fatalf("CurrentSession() error = %v", err)
	}

	if data.UserID != 9 || *data.UserName != name || *data.UserEmail != email {
		t.Fatalf("CurrentSession() = %+v, want cached values", data)
	}
}

func TestCurrentSessionFallsBackToDatabase(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	session, err := f.sessions.Create(context.Background(), sessionDto.NewSession{
		UserID:    1,
		ExpiresAt: time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	ctx := ctxkeys.WithSessionID(context.Background(), formatSessionID(session.ID))

	data, err := f.service.CurrentSession(ctx)
	if err != nil {
		t.Fatalf("CurrentSession() error = %v", err)
	}

	if data.UserID != 1 {
		t.Fatalf("UserID = %d, want 1", data.UserID)
	}
}

func TestCurrentSessionRequiresSessionID(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	if _, err := f.service.CurrentSession(context.Background()); !errors.Is(err, apperr.ErrUnauthorized) {
		t.Fatalf("CurrentSession() error = %v, want ErrUnauthorized", err)
	}
}

func TestCurrentSessionRejectsInvalidSessionID(t *testing.T) {
	t.Parallel()

	f := newFixture(t)

	ctx := ctxkeys.WithSessionID(context.Background(), "not-a-number")

	if _, err := f.service.CurrentSession(ctx); err == nil {
		t.Fatal("CurrentSession() error = nil, want error")
	}
}
