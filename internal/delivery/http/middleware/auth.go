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
	UserIDKey   contextKey = "user_id"
	UserRoleKey contextKey = "user_role"
	TokenKey    contextKey = "token"
)

func UserAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(UserRoleKey).(string)
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
		role, ok := r.Context().Value(UserRoleKey).(string)
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

func JWTAuthMiddleware(tokenBlacklist *auth.TokenBlacklist) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
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

			token := bearerToken[1]

			blacklisted, err := tokenBlacklist.IsBlacklisted(r.Context(), token)
			if err != nil {
				api.SendResponse(w, http.StatusInternalServerError, "Authentication failed", nil, "Error checking token status")
				return
			}
			if blacklisted {
				api.SendResponse(w, http.StatusUnauthorized, "Authentication failed", nil, "Token is no longer valid")
				return
			}

			userID, role, err := auth.ValidateTokenWithRole(token)
			if err != nil {
				api.SendResponse(w, http.StatusUnauthorized, "Authentication failed", nil, "Invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, UserRoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}
