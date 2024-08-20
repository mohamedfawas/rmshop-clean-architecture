package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/api"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/auth"
)

// Define custom types for context keys
type contextKey string

const (
	userIDKey   contextKey = "user_id"
	userRoleKey contextKey = "user_role"
)

func JWTAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Entering JWTAuthMiddleware")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			api.SendResponse(w, http.StatusUnauthorized, "Authentication failed", nil, "Missing authorization header")
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			api.SendResponse(w, http.StatusUnauthorized, "Authentication failed", nil, "Invalid authorization header format")
			return
		}

		//log.Printf("Validating token: %s", bearerToken[1])
		userID, role, err := auth.ValidateTokenWithRole(bearerToken[1])
		if err != nil {
			api.SendResponse(w, http.StatusUnauthorized, "Authentication failed", nil, "Invalid or expired token")
			return
		}

		if role != "admin" {
			api.SendResponse(w, http.StatusForbidden, "Access denied", nil, "Admin access required")
			return
		}

		log.Printf("Token validated successfully. UserID: %d, Role: %s", userID, role)
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		ctx = context.WithValue(ctx, userRoleKey, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func UserAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Entering UserAuthMiddleware")

		authHeader := r.Header.Get("Authorization")
		//log.Printf("Authorization header: %s", authHeader)

		if authHeader == "" {
			//log.Println("Missing authorization header")
			api.SendResponse(w, http.StatusUnauthorized, "Authentication failed", nil, "Missing authorization header")
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || !strings.EqualFold(bearerToken[0], "Bearer") {
			// log.Println("Invalid authorization header format")
			api.SendResponse(w, http.StatusUnauthorized, "Authentication failed", nil, "Invalid authorization header format")
			return
		}

		// log.Printf("Validating token: %s", bearerToken[1])
		userID, role, err := auth.ValidateTokenWithRole(bearerToken[1])
		if err != nil {
			// log.Printf("Invalid or expired token: %v", err)
			api.SendResponse(w, http.StatusUnauthorized, "Authentication failed", nil, "Invalid or expired token")
			return
		}

		// Check if the role is specifically "user"
		if role != "user" {
			// log.Printf("Token belongs to non-user role: %s", role)
			api.SendResponse(w, http.StatusForbidden, "Authentication failed", nil, "Access denied: user role required")
			return
		}

		log.Printf("Token validated successfully. UserID: %d, Role: %s", userID, role)
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		ctx = context.WithValue(ctx, userRoleKey, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
