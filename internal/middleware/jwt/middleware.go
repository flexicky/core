package jwt

import (
	"context"
	"strings"

	"core/internal/ctxkeys"
	"core/internal/service/token"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const bearerPrefix = "Bearer "

// Middleware authenticates unary gRPC calls with a bearer access token.
type Middleware struct {
	tokens    token.Service
	publicRPC map[string]struct{}
}

// New builds the interceptor. The methods listed in public are reachable
// without a token (login, register).
func New(tokens token.Service, public ...string) *Middleware {
	allowed := make(map[string]struct{}, len(public))
	for _, method := range public {
		allowed[method] = struct{}{}
	}

	return &Middleware{tokens: tokens, publicRPC: allowed}
}

// UnaryServerInterceptor rejects calls without a valid access token and stores
// the session id from the token in the request context.
func (m *Middleware) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if _, ok := m.publicRPC[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		sessionID, err := m.sessionIDFromContext(ctx)
		if err != nil {
			return nil, err
		}

		return handler(ctxkeys.WithSessionID(ctx, sessionID), req)
	}
}

func (m *Middleware) sessionIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization token")
	}

	raw := values[0]
	if !strings.HasPrefix(raw, bearerPrefix) {
		return "", status.Error(codes.Unauthenticated, "invalid authorization header format")
	}

	claims, err := m.tokens.ParseAccessToken(strings.TrimPrefix(raw, bearerPrefix))
	if err != nil {
		return "", status.Error(codes.Unauthenticated, "invalid token")
	}

	if claims.SessionID == "" {
		return "", status.Error(codes.Unauthenticated, "token has no session")
	}

	return claims.SessionID, nil
}
