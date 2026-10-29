// Package ctxkeys stores the per-request values that travel through context.
// Using a private key type prevents collisions with keys defined in other packages.
package ctxkeys

import "context"

type key int

const (
	clientInfoKey key = iota
	sessionIDKey
)

// ClientInfo describes where a request came from.
type ClientInfo struct {
	UserAgent string
	IP        string
}

// Unknown is used when the caller did not provide the information.
const unknown = "unknown"

// WithClientInfo stores client information in ctx.
func WithClientInfo(ctx context.Context, info ClientInfo) context.Context {
	if info.UserAgent == "" {
		info.UserAgent = unknown
	}
	if info.IP == "" {
		info.IP = unknown
	}

	return context.WithValue(ctx, clientInfoKey, info)
}

// ClientInfoFrom returns the client information stored in ctx.
func ClientInfoFrom(ctx context.Context) ClientInfo {
	if info, ok := ctx.Value(clientInfoKey).(ClientInfo); ok {
		return info
	}

	return ClientInfo{UserAgent: unknown, IP: unknown}
}

// WithSessionID stores the authenticated session identifier in ctx.
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey, sessionID)
}

// SessionID returns the authenticated session identifier stored in ctx.
func SessionID(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(sessionIDKey).(string)
	if !ok || sessionID == "" {
		return "", false
	}

	return sessionID, true
}
