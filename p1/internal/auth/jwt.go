package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int32  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID int32, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func GenerateRefreshToken(userID int32, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_REFRESH_SECRET")))
}

func ValidateToken(tokenStr string, isRefresh bool) (*Claims, error) {
	secret := os.Getenv("JWT_SECRET")
	if isRefresh {
		secret = os.Getenv("JWT_REFRESH_SECRET")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return token.Claims.(*Claims), nil
}
