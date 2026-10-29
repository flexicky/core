package ctxkeys

import (
	"context"
	"testing"
)

func TestClientInfoRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := WithClientInfo(context.Background(), ClientInfo{UserAgent: "grpc-go", IP: "10.0.0.1"})

	info := ClientInfoFrom(ctx)
	if info.UserAgent != "grpc-go" || info.IP != "10.0.0.1" {
		t.Fatalf("ClientInfoFrom() = %+v, want the stored values", info)
	}
}

func TestClientInfoDefaults(t *testing.T) {
	t.Parallel()

	info := ClientInfoFrom(context.Background())
	if info.UserAgent != "unknown" || info.IP != "unknown" {
		t.Fatalf("ClientInfoFrom() = %+v, want unknown values", info)
	}

	empty := ClientInfoFrom(WithClientInfo(context.Background(), ClientInfo{}))
	if empty.UserAgent != "unknown" || empty.IP != "unknown" {
		t.Fatalf("empty client info must be normalized, got %+v", empty)
	}
}

func TestSessionID(t *testing.T) {
	t.Parallel()

	if _, ok := SessionID(context.Background()); ok {
		t.Fatal("SessionID() must report a miss for a context without a session")
	}

	if _, ok := SessionID(WithSessionID(context.Background(), "")); ok {
		t.Fatal("SessionID() must treat an empty id as missing")
	}

	sessionID, ok := SessionID(WithSessionID(context.Background(), "42"))
	if !ok || sessionID != "42" {
		t.Fatalf("SessionID() = %q, %v; want 42, true", sessionID, ok)
	}
}
