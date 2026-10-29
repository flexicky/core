package utils

import (
	"context"

	"core/internal/ctxkeys"

	"google.golang.org/grpc/metadata"
)

const unknown = "unknown"

// ClientInfoFromMetadata extracts the user agent and the client IP from the
// incoming gRPC metadata.
func ClientInfoFromMetadata(ctx context.Context) ctxkeys.ClientInfo {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctxkeys.ClientInfo{UserAgent: unknown, IP: unknown}
	}

	return ctxkeys.ClientInfo{
		UserAgent: firstOrUnknown(md.Get("user-agent")),
		IP:        firstNonEmpty(md.Get("x-forwarded-for"), md.Get("x-real-ip")),
	}
}

func firstOrUnknown(values []string) string {
	if len(values) == 0 || values[0] == "" {
		return unknown
	}

	return values[0]
}

func firstNonEmpty(groups ...[]string) string {
	for _, values := range groups {
		if len(values) > 0 && values[0] != "" {
			return values[0]
		}
	}

	return unknown
}

// Ptr returns a pointer to value. It is handy for proto optional fields.
func Ptr[T any](value T) *T {
	return &value
}
