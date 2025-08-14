package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/MatTwix/Food-Delivery-Agregator/api-gateway/config"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := headerParts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
			if config.Cfg.JWT.Secret == "" {
				return []byte(""), errors.New("jwt secret is not set")
			}

			return []byte(config.Cfg.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			slog.Error("failed to confirm token", "error", err)
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-User-Id", claims.UserID)
		r.Header.Set("X-User-Role", claims.Role)
		slog.Info("token validated successfully", "user_id", claims.UserID, "role", claims.Role)
		next.ServeHTTP(w, r)
	})
}
