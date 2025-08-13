package auth

import (
	"time"

	"github.com/MatTwix/Food-Delivery-Agregator/users-service/config"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateTokens(userID, userRole string) (accessToken, refreshToken string, err error) {
	accessClaims := &Claims{
		UserID: userID,
		Role:   userRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}

	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(config.Cfg.JWT.Secret))
	if err != nil {
		return "", "", err
	}

	refreshClaims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 7 * time.Hour)),
		},
	}

	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(config.Cfg.JWT.Secret))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
