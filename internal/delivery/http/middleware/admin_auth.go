package middleware

import (
	"log"
	"net/http"
)

func AdminAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Entering AdminAuthMiddleware")
		role, ok := r.Context().Value("user_role").(string)
		if !ok || role != "admin" {
			log.Printf("Admin access required. Current role: %s", role)
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}
		log.Println("Admin access granted")
		next.ServeHTTP(w, r)
	}
}
