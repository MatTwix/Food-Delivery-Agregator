package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/common/auth"
	"github.com/go-chi/chi"
)

func Authorize(allowedRoles ...auth.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Header.Get("X-User-Role")

			isAllowed := false
			for _, role := range allowedRoles {
				if userRole == role.String() {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				http.Error(w, "Forbidden: insufficient permisions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func AuthorizeOwnerOrRoles(getOwnerID func(ctx context.Context, targetID string) (string, error), allowedRoles ...auth.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Header.Get("X-User-Id")
			userRole := r.Header.Get("X-User-Role")

			targetID := chi.URLParam(r, "id")
			ownerID, err := getOwnerID(r.Context(), targetID)
			if err != nil {
				slog.Error("failed to get owner id", "error", err)
				http.Error(w, "Failed to get owner id", http.StatusInternalServerError)
				return
			}

			isAllowed := false

			if ownerID == userID {
				isAllowed = true
			} else {
				for _, role := range allowedRoles {
					if userRole == role.String() {
						isAllowed = true
						break
					}
				}
			}

			if !isAllowed {
				http.Error(w, "Forbidden: insufficient permisions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
