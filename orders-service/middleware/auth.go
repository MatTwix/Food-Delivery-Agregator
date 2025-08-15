package middleware

import (
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/common/auth"
)

func CheckRole(requiredRole auth.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Header.Get("X-User-Role")
			if userRole != requiredRole.String() && userRole != auth.RoleAdmin.String() {
				http.Error(w, "Forbidden: insufficient permisions", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
