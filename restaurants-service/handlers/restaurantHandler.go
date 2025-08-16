package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/messaging"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/models"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/store"
	"github.com/go-chi/chi/v5"
)

type RestaurantHandler struct {
	store    *store.RestaurantStore
	producer *messaging.Producer
}

type restaurantsInput struct {
	OwnerID     string `json:"owner_id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Address     string `json:"address" validate:"required"`
	PhoneNumber string `json:"phone_number"`
}

type DeletionMessage struct {
	ID string `json:"id"`
}

func NewRestaurantHandler(s *store.RestaurantStore, p *messaging.Producer) *RestaurantHandler {
	return &RestaurantHandler{
		store:    s,
		producer: p,
	}
}

func (h *RestaurantHandler) GetRestaurants(w http.ResponseWriter, r *http.Request) {
	restaurants, err := h.store.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Error getting restaurants", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(restaurants)
}

func (h *RestaurantHandler) GetRestaurantByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	restaurant, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Error getting restaurant by ID", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(restaurant)
}

func (h *RestaurantHandler) CreateRestaurant(w http.ResponseWriter, r *http.Request) {
	var input restaurantsInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := config.Validator.Struct(&input); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	restaurant := models.Restaurant{
		OwnerID:     input.OwnerID,
		Name:        input.Name,
		Address:     input.Address,
		PhoneNumber: input.PhoneNumber,
	}

	if err := h.store.Create(r.Context(), &restaurant); err != nil {
		http.Error(w, "Error creating restaurant", http.StatusInternalServerError)
		return
	}

	eventBody, err := json.Marshal(restaurant)
	if err != nil {
		slog.Error("failed to marshal restaurant for Kafka event", "error", err)
	} else {
		h.producer.Produce(r.Context(), messaging.RestaurantCreatedTopic, []byte(restaurant.ID), eventBody)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(restaurant)
}

func (h *RestaurantHandler) UpdateRestaurant(w http.ResponseWriter, r *http.Request) {
	var input restaurantsInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := config.Validator.Struct(&input); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	id := chi.URLParam(r, "id")

	restaurant := models.Restaurant{
		ID:          id,
		Name:        input.Name,
		Address:     input.Address,
		PhoneNumber: input.PhoneNumber,
	}

	if err := h.store.Update(r.Context(), &restaurant); err != nil {
		http.Error(w, "Error updating restaurant", http.StatusInternalServerError)
		return
	}

	eventBody, err := json.Marshal(restaurant)
	if err != nil {
		slog.Error("failed to marshal restaurant for Kafka event", "error", err)
	} else {
		h.producer.Produce(r.Context(), messaging.RestaurantUpdatedTopic, []byte(restaurant.ID), eventBody)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(restaurant)
}

func (h *RestaurantHandler) DeleteRestaurant(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.store.Delete(r.Context(), id); err != nil {
		http.Error(w, "Error deleting restaurant", http.StatusInternalServerError)
		return
	}

	message := DeletionMessage{
		ID: id,
	}

	eventBody, err := json.Marshal(message)
	if err != nil {
		slog.Error("failed to marshal restaurant for Kafka event", "error", err)
	} else {
		h.producer.Produce(r.Context(), messaging.RestaurantDeletedTopic, []byte(message.ID), eventBody)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Restaurant deleted!")
}
