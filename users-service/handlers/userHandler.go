package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/users-service/auth"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/models"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/store"
	"github.com/jackc/pgx/v5"
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

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponce struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
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

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := config.Validator.Struct(&req); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.store.GetByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		} else {
			slog.Error("failed to get user by email", "error", err)
			http.Error(w, "Error getting user by email", http.StatusInternalServerError)
		}
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	accessToken, refreshToken, err := auth.GenerateTokens(user.ID, user.Role)
	if err != nil {
		slog.Error("failed to generate tokens", "error", err)
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponce{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *UserHandler) GetAllUses(w http.ResponseWriter, r *http.Request) {
	users, err := h.store.GetAll(r.Context())
	if err != nil {
		slog.Error("failed to get all users", "error", err)
		http.Error(w, "Error getting all users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
