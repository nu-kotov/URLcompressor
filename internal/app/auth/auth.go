package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/golang-jwt/jwt/v5"
)

// Claims - структура - части JWT-токена.
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

// TokenExp - время жизни токена.
const TokenExp = time.Hour * 3

// SecretKey - секретный ключ для подписи токена.
const SecretKey = "supersecretkey"

// BuildJWTString создает JWT токен.
func BuildJWTString() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: uuid.New().String(),
	})

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetUserID получает ID пользователя из JWT токена.
func GetUserID(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(SecretKey), nil
		})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("token is not valid")
	}

	return claims.UserID, nil
}
