package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	commonAuth "github.com/MatTwix/Food-Delivery-Agregator/common/auth"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/auth"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/messaging"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/models"
	"github.com/MatTwix/Food-Delivery-Agregator/users-service/store"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type UserHandler struct {
	userStore  *store.UserStore
	tokenStore *store.TokenStore
	producer   *messaging.Producer
}

func NewUserHandler(uStore *store.UserStore, tStore *store.TokenStore, producer *messaging.Producer) *UserHandler {
	return &UserHandler{
		userStore:  uStore,
		tokenStore: tStore,
		producer:   producer,
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

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LoginResponce struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ChangeRoleRequest struct {
	Role string `json:"role" validate:"required,min=2,max=15"`
	Name string `json:"name" validate:"required,min=2,max=50"`
}

type ChangeRoleResponce struct {
	UserID   string `json:"user_id"`
	PrevRole string `json:"prev_role"`
	NewRole  string `json:"new_role"`
	Name     string `json:"name"`
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

	if err := h.userStore.Create(r.Context(), user); err != nil {
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

	user, err := h.userStore.GetByEmail(r.Context(), req.Email)
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

	accessToken, refreshToken, refreshExpiresAt, err := auth.GenerateTokens(user.ID, user.Role)
	if err != nil {
		slog.Error("failed to generate tokens", "error", err)
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	token := models.Token{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: refreshExpiresAt,
	}

	if err := h.tokenStore.SaveRefreshToken(r.Context(), &token); err != nil {
		slog.Error("failed to save token to db", "error", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponce{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *UserHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	claims, err := auth.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		slog.Error("failed to validate token", "token", req.RefreshToken, "error", err)
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	//TODO: check that token still exists in db

	user, err := h.userStore.GetByID(r.Context(), claims.UserID)
	if err != nil {
		slog.Error("failed to get user by id", "user_id", claims.UserID, "error", err)
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	h.tokenStore.DeleteRefreshToken(r.Context(), req.RefreshToken)

	accessToken, newRefreshToken, newRefreshExpiresAt, err := auth.GenerateTokens(user.ID, user.Role)
	if err != nil {
		slog.Error("failed to generate tokens", "user_id", user.ID, "error", err)
		http.Error(w, "Failed to generate tokens", http.StatusUnauthorized)
		return
	}

	token := models.Token{
		UserID:    user.ID,
		Token:     newRefreshToken,
		ExpiresAt: newRefreshExpiresAt,
	}

	if err := h.tokenStore.SaveRefreshToken(r.Context(), &token); err != nil {
		slog.Error("failed to save token", "token", newRefreshToken, "error", err)
		http.Error(w, "Failed to save new session", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponce{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	})
}

func (h *UserHandler) GetAllUses(w http.ResponseWriter, r *http.Request) {
	users, err := h.userStore.GetAll(r.Context())
	if err != nil {
		slog.Error("failed to get all users", "error", err)
		http.Error(w, "Error getting all users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) ChangeUserRole(w http.ResponseWriter, r *http.Request) {
	var req ChangeRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := config.Validator.Struct(&req); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	id := chi.URLParam(r, "id")

	prevRole, err := h.userStore.ChangeRole(r.Context(), id, commonAuth.Role(req.Role))

	if err != nil {
		slog.Error("failed to change user role", "user_id", id, "error", err)
		http.Error(w, "Failed to change user role", http.StatusInternalServerError)
		return
	}

	event := ChangeRoleResponce{
		UserID:   id,
		PrevRole: prevRole,
		NewRole:  req.Role,
		Name:     req.Name,
	}

	eventBody, err := json.Marshal(event)
	if err != nil {
		slog.Error("failed to marshal message for Kafka event", "error", err)
		http.Error(w, "Error sending role change event", http.StatusInternalServerError)
		return
	} else {
		h.producer.Produce(r.Context(), messaging.UsersRoleAssignedTopic, []byte(id), eventBody)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(event)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.userStore.Delete(r.Context(), id)
	if err != nil {
		slog.Error("failed to delete user", "error", err)
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("User successfully deleted!")
}
