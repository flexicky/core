package token

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	SessionId string `json:"session_id"`
	jwt.RegisteredClaims
}
