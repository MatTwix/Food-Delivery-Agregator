package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/messaging"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/models"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/store"
)

type RestaurantHandler struct {
	store    *store.RestaurantStore
	producer *messaging.Producer
}

type restaurantsInput struct {
	Name        string `json:"name" validate:"required"`
	Address     string `json:"address" validate:"required"`
	PhoneNumber string `json:"phone_number"`
}

func NewRestaurantHandler(s *store.RestaurantStore, p *messaging.Producer) *RestaurantHandler {
	return &RestaurantHandler{
		store:    s,
		producer: p,
	}
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
		log.Printf("Error marshaling restaurant for Kafka event: %v", err)
	} else {
		h.producer.Produce(r.Context(), []byte(restaurant.ID), eventBody)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(restaurant)
}
