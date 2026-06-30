package auth

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("недействительный токен")

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-dev-secret-change-me-in-production-32chars!!"
	}
	return []byte(secret)
}

func getExpirationHours() int {
	hoursStr := os.Getenv("JWT_EXPIRATION_HOURS")
	if hoursStr == "" {
		return 24
	}
	hours, err := strconv.Atoi(hoursStr)
	if err != nil {
		return 24
	}
	return hours
}

func GenerateToken(userID int64) (string, error) {
	expirationHours := getExpirationHours()
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "volunteer_platform",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

func ValidateToken(tokenString string) (int64, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return getJWTSecret(), nil
	})

	if err != nil || !token.Valid {
		return 0, ErrInvalidToken
	}

	return claims.UserID, nil
}
