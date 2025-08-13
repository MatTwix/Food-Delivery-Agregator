package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/users-service/auth"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/models"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/store"
)

type UserHandler struct {
	store *store.UserStore
}

func NewUserHandler(s *store.UserStore) *UserHandler {
	return &UserHandler{
		store: s,
	}
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := config.Validator.Struct(&req); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		slog.Error("failed to hash password", "error", err)
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	if err := h.store.Create(r.Context(), user); err != nil {
		//TODO: process repeating email error
		slog.Error("failed to create user", "error", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	user.PasswordHash = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
