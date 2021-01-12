package store

import "github.com/dgrijalva/jwt-go"

// JWTClaims is used to generate and verify JWT security tokens
type JWTClaims struct {
	Username string
	jwt.StandardClaims
}
