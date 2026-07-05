package utils

import (
	"errors"
	"fmt"
	"time"

	"task-system/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	UserID    string `json:"UserID"`
	SessionID string `json:"SessionID"`
	Email     string `json:"Email"`
	jwt.RegisteredClaims
}

func getJWTKey() []byte {
	return []byte(config.AppConfig.JWTSecret)
}

// GenerateAccessToken generates a short-lived access token valid for 15 minutes
func GenerateAccessToken(userID string, sessionID string, email string) (string, error) {
	expirationTime := time.Now().Add(15 * time.Minute)

	claims := &TokenClaims{
		UserID:    userID,
		SessionID: sessionID,
		Email:     email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTKey())
}

// ValidateAccessToken decrypts and verifies the JWT token signature
func ValidateAccessToken(tokenStr string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure token uses HMAC (HS256) – reject any other algorithm (including "none")
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return getJWTKey(), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
