package utils

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func ExtractClientInfo(ctx context.Context) (userAgent string, ipAddress string) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return "unknown", "unknown"
	}

	if userAgents := md.Get("user-agent"); len(userAgents) > 0 {
		userAgent = userAgents[0]
	} else {
		userAgent = "unknown"
	}

	if ips := md.Get("x-forwarded-for"); len(ips) > 0 {
		ipAddress = ips[0]
	} else if ips := md.Get("x-real-ip"); len(ips) > 0 {
		ipAddress = ips[0]
	} else {
		ipAddress = "unknown"
	}

	return userAgent, ipAddress
}
