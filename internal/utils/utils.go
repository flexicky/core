package utils

import (
	"context"
	"time"

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

func Process(ctx context.Context, timeSleep time.Duration, maxRetries int, fn func(ctx context.Context)) {
	if timeSleep == 0 {
		timeSleep = time.Second
	}

	ticker := time.NewTicker(timeSleep)
	defer ticker.Stop()

	run := func() {
		defer func() {
			if r := recover(); r != nil {
				// TODO позже придумавть как отлавливать в логи или пробрасывать
			}
		}()
		fn(ctx)
	}

	run()
	retries := 1

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if maxRetries > 0 && retries >= maxRetries {
				return
			}
			run()
			retries++
		}
	}
}
