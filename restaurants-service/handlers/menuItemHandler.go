package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/models"
	"github.com/MatTwix/Food-Delivery-Agregator/restaurants-service/store"
	"github.com/go-chi/chi/v5"
)

type MenuItemHandler struct {
	store *store.MenuItemStore
}

type MenuItemInputCreate struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`
	Price       int    `json:"price" validate:"required"`
}

type MenuItemInputUpdate struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`
	Price       int    `json:"price" validate:"required"`
}

func NewMenuItemHandler(s *store.MenuItemStore) *MenuItemHandler {
	return &MenuItemHandler{
		store: s,
	}
}

func (h *MenuItemHandler) GetMenuItems(w http.ResponseWriter, r *http.Request) {
	menuItems, err := h.store.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Error getting menu items", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(menuItems)
}

func (h *MenuItemHandler) GetMenuItemsByRestaurantID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	menuItems, err := h.store.GetByRestaurantID(r.Context(), id)
	if err != nil {
		http.Error(w, "Error getting menu items", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(menuItems)
}

func (h *MenuItemHandler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	var input MenuItemInputCreate

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := config.Validator.Struct(&input); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	restaurantID := chi.URLParam(r, "id")

	menuItem := models.MenuItem{
		RestaurantID: restaurantID,
		Name:         input.Name,
		Description:  input.Description,
		Price:        input.Price,
	}

	if err := h.store.Create(r.Context(), &menuItem); err != nil {
		http.Error(w, "Error creating menu item", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(menuItem)
}

func (h *MenuItemHandler) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	var input MenuItemInputUpdate

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := config.Validator.Struct(&input); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	id := chi.URLParam(r, "id")

	menuItem := models.MenuItem{
		ID:          id,
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
	}

	if err := h.store.Update(r.Context(), &menuItem); err != nil {
		http.Error(w, "Error updating menu item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(menuItem)
}

func (h *MenuItemHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.store.Delete(r.Context(), id); err != nil {
		http.Error(w, "Error deleting restaurant", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Menu item deleted!")
}
