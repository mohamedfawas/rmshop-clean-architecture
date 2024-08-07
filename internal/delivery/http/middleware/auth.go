package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/auth"
)

func JWTAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Entering JWTAuthMiddleware")
		log.Printf("Request Headers: %+v", r.Header)

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("Missing authorization header")
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		log.Printf("Authorization header: %s", authHeader)

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			log.Println("Invalid authorization header format")
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		userID, role, err := auth.ValidateTokenWithRole(bearerToken[1])
		if err != nil {
			log.Printf("Invalid or expired token: %v", err)
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		log.Printf("Token validated. UserID: %d, Role: %s", userID, role)
		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "user_role", role)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
