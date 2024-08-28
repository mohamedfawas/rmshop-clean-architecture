// Package auth provides JWT (JSON Web Token) authentication utilities.
package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

// jwtSecret is the secret key used for signing JWTs
var JWTSecret []byte

// InitJWTSecret initializes the JWT secret from the provided configuration file
func InitJWTSecret(secret string) {
	JWTSecret = []byte(secret)
}

// GenerateTokenWithRole generates a JWT token with a specified role for a given user
//
// Parameters:
// - userID: The unique identifier of the user
// - role: The role of the user (e.g., "admin", "user")
//
// Returns:
// - string: The signed JWT token as a string
// - error: An error if signing the token fails
func GenerateTokenWithRole(userID int64, role string) (string, error) {
	// Define the claims for the JWT token, including the user ID, role, and expiration time
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().UTC().Add(time.Hour * 24).Unix(), // Token expires in 24 hours.
	}

	// Create a new token with the specified signing method and claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token using the secret key and return it
	return token.SignedString(JWTSecret)
}

// parseToken parses a JWT token string and returns the parsed token if valid.
//
// Parameters:
// - tokenString: The JWT token as a string.
//
// Returns:
// - *jwt.Token: A pointer to the parsed JWT token object.
// - error: An error if the token parsing fails or if the signing method is unexpected.
func parseToken(tokenString string) (*jwt.Token, error) {
	// Parse the JWT token string, validating the signing method and returning the secret key
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check if the signing method is HMAC (expected method). If not, return an error
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, utils.ErrUnexpectedSigning // Return an error if the signing method is not as expected
		}
		// Return the secret key used to sign the token
		return JWTSecret, nil
	})
}

// GetClaimsFromToken extracts the claims from a JWT token string.
//
// Parameters:
// - tokenString: The JWT token as a string.
//
// Returns:
// - jwt.MapClaims: A map containing the token's claims if the token is valid.
// - error: An error if the token is invalid or if extracting the claims fails.
func GetClaimsFromToken(tokenString string) (jwt.MapClaims, error) {
	// Parse the token string using the parseToken function
	token, err := parseToken(tokenString)
	if err != nil {
		// Return an error if token parsing fails
		return nil, err
	}
	// Check if the token's claims can be asserted as jwt.MapClaims and if the token is valid
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Return the claims if the token is valid
		return claims, nil
	}

	// Return an error if the token is invalid or the claims cannot be extracted
	return nil, utils.ErrInvalidToken
}

// ValidateToken validates a JWT token and extracts the user ID from it.
//
// Parameters:
// - tokenString: The JWT token as a string.
//
// Returns:
// - int64: The user ID extracted from the token if it is valid.
// - error: An error if the token is invalid or if the user ID claim cannot be extracted.
func ValidateToken(tokenString string) (int64, error) {
	// Extract the claims from the token using the GetClaimsFromToken function
	claims, err := GetClaimsFromToken(tokenString)
	if err != nil {
		// Return an error if the claims cannot be extracted or the token is invalid
		return 0, err
	}

	// Attempt to extract the "user_id" claim and convert it to a float64
	userID, ok := claims["user_id"].(float64)
	if !ok {
		// Return an error if the "user_id" claim is missing or of an unexpected type
		return 0, utils.ErrInvalidUserID
	}

	// Return the user ID as an int64
	return int64(userID), nil
}

// ValidateTokenWithRole validates a JWT token and extracts both the user ID and role from it.
//
// Parameters:
// - tokenString: The JWT token as a string.
//
// Returns:
// - int64: The user ID extracted from the token if it is valid.
// - string: The role associated with the user extracted from the token.
// - error: An error if the token is invalid, or if the user ID or role claims cannot be extracted.
func ValidateTokenWithRole(tokenString string) (int64, string, error) {
	// Extract the claims from the token using the GetClaimsFromToken function
	claims, err := GetClaimsFromToken(tokenString)
	if err != nil {
		// Return an error if the claims cannot be extracted or the token is invalid
		return 0, "", err
	}

	// Attempt to extract the "user_id" claim and convert it to a float64
	userID, ok := claims["user_id"].(float64)
	if !ok {
		// Return an error if the "user_id" claim is missing or of an unexpected type
		return 0, "", utils.ErrInvalidUserID
	}

	// Attempt to extract the "role" claim and ensure it is a string
	role, ok := claims["role"].(string)
	if !ok {
		// Return an error if the "role" claim is missing or of an unexpected type
		return 0, "", utils.ErrInvalidRole
	}

	// Return the user ID and role
	return int64(userID), role, nil
}

// RefreshToken refreshes a JWT token by updating its expiration time and generating a new token.
//
// Parameters:
// - tokenString: The JWT token as a string.
//
// Returns:
// - string: The refreshed JWT token with an updated expiration time.
// - error: An error if the token is invalid or if generating the new token fails.
func RefreshToken(tokenString string) (string, error) {
	// Extract the claims from the token using the GetClaimsFromToken function
	claims, err := GetClaimsFromToken(tokenString)
	if err != nil {
		// Return an error if the claims cannot be extracted or the token is invalid
		return "", err
	}

	// Update the "exp" claim to extend the token's expiration time by 24 hours
	claims["exp"] = time.Now().UTC().Add(time.Hour * 24).Unix()

	// Create a new token with the updated claims and sign it using the secret key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Return the signed, refreshed token
	return token.SignedString(JWTSecret)
}
