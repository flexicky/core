package utils

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestClientInfoFromMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		md        metadata.MD
		userAgent string
		ip        string
	}{
		{
			name:      "no metadata",
			md:        nil,
			userAgent: unknown,
			ip:        unknown,
		},
		{
			name:      "user-agent and x-forwarded-for",
			md:        metadata.Pairs("user-agent", "grpc-go/1.83", "x-forwarded-for", "10.0.0.1"),
			userAgent: "grpc-go/1.83",
			ip:        "10.0.0.1",
		},
		{
			name:      "x-real-ip fallback",
			md:        metadata.Pairs("x-real-ip", "192.168.1.5"),
			userAgent: unknown,
			ip:        "192.168.1.5",
		},
		{
			name:      "empty values fall back to unknown",
			md:        metadata.Pairs("user-agent", "", "x-forwarded-for", ""),
			userAgent: unknown,
			ip:        unknown,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			if test.md != nil {
				ctx = metadata.NewIncomingContext(ctx, test.md)
			}

			info := ClientInfoFromMetadata(ctx)
			if info.UserAgent != test.userAgent {
				t.Fatalf("UserAgent = %q, want %q", info.UserAgent, test.userAgent)
			}

			if info.IP != test.ip {
				t.Fatalf("IP = %q, want %q", info.IP, test.ip)
			}
		})
	}
}

func TestPtr(t *testing.T) {
	t.Parallel()

	value := Ptr("core")
	if value == nil || *value != "core" {
		t.Fatalf("Ptr() = %v, want pointer to %q", value, "core")
	}
}
