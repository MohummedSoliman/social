package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedBasicError(w, r, fmt.Errorf("authorization header is messing"))
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) < 2 || parts[0] != "Basic" {
				app.unauthorizedBasicError(w, r, fmt.Errorf("authorization header is malformed"))
				return
			}

			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unauthorizedBasicError(w, r, err)
				return
			}

			username := app.config.auth.basic.user
			pass := app.config.auth.basic.pass

			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 || creds[0] != username || creds[1] != pass {
				app.unauthorizedError(w, r, fmt.Errorf("invalid credentials"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (app *application) AuthTokenMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				app.unauthorizedError(w, r, fmt.Errorf("authorization header is messing"))
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				app.unauthorizedError(w, r, fmt.Errorf("authorization header is malformed"))
				return
			}

			token := parts[1]
			jwtToken, err := app.authenticator.ValidateToken(token)
			if err != nil {
				app.unauthorizedError(w, r, err)
				return
			}

			claims, _ := jwtToken.Claims.(jwt.MapClaims)
			userID, err := strconv.Atoi(fmt.Sprintf("%.f", claims["sub"]))
			if err != nil {
				app.unauthorizedError(w, r, err)
				return
			}

			ctx := r.Context()
			user, err := app.store.Users.GetUserByID(ctx, int64(userID))
			if err != nil {
				app.unauthorizedError(w, r, err)
				return
			}

			ctx = context.WithValue(ctx, USERKEY, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
