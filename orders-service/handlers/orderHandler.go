package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	pb "github.com/MatTwix/Food-Delivery-Agregator/common/proto"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/models"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/store"
)

type CreateOrderRequest struct {
	RestaurantID string `json:"restaurant_id" validate:"required"`
	Items        []struct {
		MenuItemID string `json:"menu_item_id" validate:"required"`
		Quantity   int    `json:"quantity" validate:"required"`
	} `json:"items"`
}

type OrderHandler struct {
	store           *store.OrderStore
	restaurantStore *store.RestaurantStore
	grpcClient      pb.RestaurantServiceClient
}

func NewOrderHandler(os *store.OrderStore, rs *store.RestaurantStore, grpc pb.RestaurantServiceClient) *OrderHandler {
	return &OrderHandler{
		store:           os,
		restaurantStore: rs,
		grpcClient:      grpc,
	}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := config.Validator.Struct(&req); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Check req.RestaurantID existanse

	menuItemIDs := []string{}
	for _, item := range req.Items {
		menuItemIDs = append(menuItemIDs, item.MenuItemID)
	}

	grpcReq := &pb.GetMenuItemsRequest{
		RestaurantId: req.RestaurantID,
		MenuItemIds:  menuItemIDs,
	}

	grpcRes, err := h.grpcClient.GetMenuItems(r.Context(), grpcReq)
	if err != nil {
		log.Printf("Error call to restaurants-service via gRPC: %v", err)
		http.Error(w, "Error getting menu items info", http.StatusInternalServerError)
		return
	}

	if len(grpcRes.MenuItems) != len(menuItemIDs) {
		http.Error(w, "Some menu items could not be found", http.StatusBadRequest)
		return
	}

	order := &models.Order{
		RestaurantID: req.RestaurantID,
		Status:       "pending",
	}

	totalPrice := 0.0
	menuItemsMap := make(map[string]*pb.MenuItem)
	for _, item := range grpcRes.MenuItems {
		menuItemsMap[item.Id] = item
	}

	for _, reqItem := range req.Items {
		menuItem, ok := menuItemsMap[reqItem.MenuItemID]
		if !ok {
			http.Error(w, fmt.Sprintf("Menu item %s not found after gRPC call", reqItem.MenuItemID), http.StatusInternalServerError)
			return
		}
		order.Items = append(order.Items, models.OrderItem{
			MenuItemID: reqItem.MenuItemID,
			Quantity:   reqItem.Quantity,
			Price:      menuItem.Price,
		})
		totalPrice += menuItem.Price * float64(reqItem.Quantity)
	}
	order.TotalPrice = totalPrice

	if err := h.store.Create(r.Context(), order); err != nil {
		log.Printf("Error creating order: %v", err)
		http.Error(w, "Error creating order", http.StatusInternalServerError)
		return
	}

	// TODO: send order.created event to kafka

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}
