package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/models"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/store"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5"
)

type CourierHandler struct {
	store *store.CourierStore
}

type CourierInputCreate struct {
	Name string `json:"name" validate:"required"`
}

type CourierInputUpdate struct {
	Name   string `json:"string" validate:"required"`
	Status string `json:"status" validate:"required"`
}

func NewCourierHandler(s *store.CourierStore) *CourierHandler {
	return &CourierHandler{
		store: s,
	}
}

func (h *CourierHandler) GetCouriers(w http.ResponseWriter, r *http.Request) {
	couriers, err := h.store.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Error getting couriers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(couriers)
}

func (h *CourierHandler) GetAvailableCourier(w http.ResponseWriter, r *http.Request) {
	courier, err := h.store.GetAvailable(r.Context())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "No available couriers found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error getting available courier", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(courier)
}

func (h *CourierHandler) CreateCourier(w http.ResponseWriter, r *http.Request) {
	var input CourierInputCreate

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := config.Validator.Struct(&input); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	courier := models.Courier{
		Name: input.Name,
	}

	if err := h.store.Create(r.Context(), &courier); err != nil {
		http.Error(w, "Error creating courier", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(courier)
}

func (h *CourierHandler) UpdateCourier(w http.ResponseWriter, r *http.Request) {
	var input CourierInputUpdate

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := config.Validator.Struct(&input); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	id := chi.URLParam(r, "id")

	courier := models.Courier{
		ID:     id,
		Name:   input.Name,
		Status: input.Status,
	}

	if err := h.store.Create(r.Context(), &courier); err != nil {
		http.Error(w, "Error creating courier", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(courier)
}

func (h *CourierHandler) DeleteCourier(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.store.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, "Error deleting couirier", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Courier deleted!")
}
