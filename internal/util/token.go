package util

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const tokenExp = time.Hour * 3

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func IsTokenValid(tokenString, secretKey string) bool {
	token, err := jwt.Parse(tokenString,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})

	if err != nil {
		return false
	}

	if !token.Valid {
		return false
	}
	return true
}

func CreateToken(userID, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("create signed jwt: %w", err)
	}

	return tokenString, nil
}

func GetUserID(tokenString, secretKey string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		return "", fmt.Errorf("jwt parse: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("token is invalid: %w", err)
	}

	return claims.UserID, nil
}

func GenerateID() (string, error) {
	b := make([]byte, 6)

	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("rand read: %w", err)
	}

	return hex.EncodeToString(b), nil
}
