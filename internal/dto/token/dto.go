package token

import "github.com/golang-jwt/jwt/v5"

// Claims is the payload of an access token.
type Claims struct {
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}
