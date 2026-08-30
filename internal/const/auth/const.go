package auth

type AuthType string

const (
	EmailAuth    AuthType = "email"
	TelegramAuth AuthType = "telegram"
	MaxAuth      AuthType = "max"
)
