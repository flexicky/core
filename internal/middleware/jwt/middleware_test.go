package jwt

import (
	"context"
	"testing"
	"time"

	"core/internal/ctxkeys"
	"core/internal/service/token"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func newTokens(t *testing.T) token.Service {
	t.Helper()

	privateKey, publicKey, err := token.GenerateKeys()
	if err != nil {
		t.Fatalf("GenerateKeys() error = %v", err)
	}

	return token.New(privateKey, publicKey, time.Hour)
}

func callInterceptor(t *testing.T, mw *Middleware, ctx context.Context, method string) (context.Context, error) {
	t.Helper()

	var captured context.Context

	_, err := mw.UnaryServerInterceptor()(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: method},
		func(ctx context.Context, _ any) (any, error) {
			captured = ctx

			return nil, nil
		},
	)

	return captured, err
}

func TestPublicMethodSkipsAuthentication(t *testing.T) {
	t.Parallel()

	mw := New(newTokens(t), "/auth.Auth/Login")

	if _, err := callInterceptor(t, mw, context.Background(), "/auth.Auth/Login"); err != nil {
		t.Fatalf("interceptor error = %v, want nil", err)
	}
}

func TestProtectedMethodRequiresMetadata(t *testing.T) {
	t.Parallel()

	mw := New(newTokens(t), "/auth.Auth/Login")

	if _, err := callInterceptor(t, mw, context.Background(), "/auth.Auth/Me"); err == nil {
		t.Fatal("interceptor error = nil, want unauthenticated")
	}
}

func TestProtectedMethodRequiresBearerPrefix(t *testing.T) {
	t.Parallel()

	tokens := newTokens(t)
	raw, _, err := tokens.CreateAccessToken(1, 2)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}

	mw := New(tokens)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", raw))

	if _, err := callInterceptor(t, mw, ctx, "/auth.Auth/Me"); err == nil {
		t.Fatal("interceptor error = nil, want unauthenticated")
	}
}

func TestProtectedMethodPropagatesSessionID(t *testing.T) {
	t.Parallel()

	tokens := newTokens(t)
	raw, _, err := tokens.CreateAccessToken(1, 99)
	if err != nil {
		t.Fatalf("CreateAccessToken() error = %v", err)
	}

	mw := New(tokens)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+raw))

	captured, err := callInterceptor(t, mw, ctx, "/auth.Auth/Me")
	if err != nil {
		t.Fatalf("interceptor error = %v", err)
	}

	sessionID, ok := ctxkeys.SessionID(captured)
	if !ok {
		t.Fatal("session id is missing in the handler context")
	}

	if sessionID != "99" {
		t.Fatalf("session id = %q, want %q", sessionID, "99")
	}
}

func TestInvalidTokenIsRejected(t *testing.T) {
	t.Parallel()

	mw := New(newTokens(t))

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer nope"))

	if _, err := callInterceptor(t, mw, ctx, "/auth.Auth/Me"); err == nil {
		t.Fatal("interceptor error = nil, want unauthenticated")
	}
}
