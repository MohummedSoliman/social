// Package auth an abstracte way for generating token
package auth

import "github.com/golang-jwt/jwt"

type Authenticator interface {
	GenerateToken(claims jwt.Claims) (string, error)
	ValidateToken(token string) (*jwt.Token, error)
}
