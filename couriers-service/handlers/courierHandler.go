package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/messaging"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/models"
	"github.com/MatTwix/Food-Delivery-Agregator/couriers-service/store"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5"
)

type CourierHandler struct {
	store    *store.CourierStore
	producer *messaging.Producer
}

type CourierUpdateRequest struct {
	Name   string `json:"name" validate:"required"`
	Status string `json:"status" validate:"required"`
}

func NewCourierHandler(s *store.CourierStore, p *messaging.Producer) *CourierHandler {
	return &CourierHandler{
		store:    s,
		producer: p,
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

func (h *CourierHandler) UpdateCourier(w http.ResponseWriter, r *http.Request) {
	var input CourierUpdateRequest

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

	if err := h.store.Update(r.Context(), &courier); err != nil {
		http.Error(w, "Error updating courier", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(courier)
}

// func (h *CourierHandler) DeleteCourier(w http.ResponseWriter, r *http.Request) {
// 	id := chi.URLParam(r, "id")
// 	err := h.store.Delete(r.Context(), id)
// 	if err != nil {
// 		http.Error(w, "Error deleting couirier", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode("Courier deleted!")
// }

func (h *CourierHandler) PickUpOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderID")
	courierID := r.Header.Get("X-User-Id")

	event := messaging.OrderPickedUpEvent{
		CourierID: courierID,
		OrderID:   orderID,
	}

	eventBody, err := json.Marshal(event)
	if err != nil {
		slog.Error("failed to marshal message for Kafka event", "error", err)
		http.Error(w, "Error marking order as picked up", http.StatusInternalServerError)
		return
	} else {
		h.producer.Produce(r.Context(), messaging.OrderPickedUpTopic, []byte(orderID), eventBody)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Order successfully marked as picked up!")
}

func (h *CourierHandler) DeliverOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "orderID")
	courierID := r.Header.Get("X-User-Id")

	event := messaging.OrderDeliveredEvent{
		CourierID: courierID,
		OrderID:   orderID,
	}

	eventBody, err := json.Marshal(event)
	if err != nil {
		slog.Error("failed to marshal message for Kafka event", "error", err)
		http.Error(w, "Error marking order as delivered", http.StatusInternalServerError)
		return
	} else {
		h.producer.Produce(r.Context(), messaging.OrderDeliveredTopic, []byte(orderID), eventBody)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Order successfully marked as delivered!")
}
