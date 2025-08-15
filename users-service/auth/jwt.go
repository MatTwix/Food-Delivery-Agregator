package auth

import (
	"errors"
	"time"

	"github.com/MatTwix/Food-Delivery-Agregator/users-service/config"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateTokens(userID, userRole string) (accessToken, refreshToken string, refreshExpiresAt time.Time, err error) {
	accessClaims := &Claims{
		UserID: userID,
		Role:   userRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}

	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(config.Cfg.JWT.Secret))
	if err != nil {
		return "", "", time.Time{}, err
	}

	refreshExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	refreshClaims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
		},
	}

	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(config.Cfg.JWT.Secret))
	if err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken, refreshToken, refreshExpiresAt, nil
}

func ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if config.Cfg.JWT.Secret == "" {
			return "", errors.New("jwt token is not set up")
		}
		return []byte(config.Cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
