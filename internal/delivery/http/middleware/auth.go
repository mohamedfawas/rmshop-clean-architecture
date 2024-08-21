package middleware

import (
	"context"
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
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			api.SendResponse(w, http.StatusUnauthorized, "Authentication failed", nil, "Missing authorization header")
			return
		}
		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || !strings.EqualFold(bearerToken[0], "Bearer") {
			api.SendResponse(w, http.StatusUnauthorized, "Authentication failed", nil, "Invalid authorization header format")
			return
		}

		userID, role, err := auth.ValidateTokenWithRole(bearerToken[1])
		if err != nil {
			api.SendResponse(w, http.StatusUnauthorized, "Authentication failed", nil, "Invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		ctx = context.WithValue(ctx, userRoleKey, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func UserAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(userRoleKey).(string)
		if !ok || role != "user" {
			api.SendResponse(w, http.StatusForbidden, "Authentication failed", nil, "Access denied: user role required")
			return
		}
		next.ServeHTTP(w, r)
	}
}

func AdminAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//log.Println("Entering AdminAuthMiddleware")

		// Retrieve the role from the context
		role, ok := r.Context().Value(userRoleKey).(string)
		if !ok {
			//log.Printf("Role not found in context")
			api.SendResponse(w, http.StatusForbidden, "Access denied", nil, "Admin access required")
			return
		}

		// Check if the role is specifically "admin"
		if role != "admin" {
			//log.Printf("Admin access required. Current role: %s", role)
			api.SendResponse(w, http.StatusForbidden, "Access denied", nil, "Admin access required")
			return
		}

		//log.Println("Admin access granted")
		next.ServeHTTP(w, r)
	}
}
