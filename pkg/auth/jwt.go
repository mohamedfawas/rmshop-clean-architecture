package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your-secret-key") // In production, this should be securely stored and not hardcoded

func GenerateToken(userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     float64(time.Now().Add(time.Hour * 24).Unix()), // Store as float64
	})

	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(float64)
		if !ok {
			return 0, errors.New("invalid user_id claim")
		}
		return int64(userID), nil
	}

	return 0, errors.New("invalid token")
}

func GetTokenClaims(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func GenerateTokenWithRole(userID int64, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString(jwtSecret)
}

// RefreshToken generates a new token with a renewed expiration time
func RefreshToken(tokenString string) (string, error) {
	claims, err := GetTokenClaims(tokenString)
	if err != nil {
		return "", err
	}

	// Update the expiration time
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateTokenWithRole(tokenString string) (int64, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return 0, "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(float64)
		if !ok {
			return 0, "", errors.New("invalid user_id claim")
		}
		role, ok := claims["role"].(string)
		if !ok {
			return 0, "", errors.New("invalid role claim")
		}
		return int64(userID), role, nil
	}

	return 0, "", errors.New("invalid token")
}
