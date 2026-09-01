package jwt

import (
	"context"
	"core/internal/service/token"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type JWTMiddleware struct {
	tokenService *token.TokenService
	whiteList    map[string]bool
}

func NewJWTMiddleware(tokenService *token.TokenService, whiteList ...string) *JWTMiddleware {
	whiteMap := make(map[string]bool)
	for _, method := range whiteList {
		whiteMap[method] = true
	}
	return &JWTMiddleware{tokenService: tokenService, whiteList: whiteMap}
}

func (m *JWTMiddleware) JWTInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		if m.whiteList[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization token")
		}

		fullToken := authHeader[0]
		if !strings.HasPrefix(fullToken, "Bearer ") {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
		}

		tokenString := strings.TrimPrefix(fullToken, "Bearer ")

		claims, err := m.tokenService.ParseAccessToken(tokenString)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = context.WithValue(ctx, "session_id", claims.SessionId)
		ctx = context.WithValue(ctx, "claims", claims)

		return handler(ctx, req)
	}
}
