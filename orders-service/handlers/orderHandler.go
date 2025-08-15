package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/MatTwix/Food-Delivery-Agregator/common/auth"
	pb "github.com/MatTwix/Food-Delivery-Agregator/common/proto"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/config"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/messaging"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/models"
	"github.com/MatTwix/Food-Delivery-Agregator/orders-service/store"
	"github.com/go-chi/chi/v5"
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
	producer        *messaging.Producer
}

func NewOrderHandler(os *store.OrderStore, rs *store.RestaurantStore, grpc pb.RestaurantServiceClient, p *messaging.Producer) *OrderHandler {
	return &OrderHandler{
		store:           os,
		restaurantStore: rs,
		grpcClient:      grpc,
		producer:        p,
	}
}

func (h *OrderHandler) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	requestingUserRole := r.Header.Get("X-User-Role")

	orders, err := h.store.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Error getting orders", http.StatusInternalServerError)
		slog.Error("failed to get orders", "error", err)
		return
	}

	//TODO: make list of verified roles and range by it
	if auth.Role(requestingUserRole) != auth.RoleAdmin && auth.Role(requestingUserRole) != auth.RoleManager {
		http.Error(w, "Forbidden: You do not have permisions to view orders list", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	requestingUserID := r.Header.Get("X-User-Id")
	requestingUserRole := r.Header.Get("X-User-Role")

	if requestingUserID == "" {
		http.Error(w, "User ID is missing", http.StatusUnauthorized)
		return
	}

	order, err := h.store.GetByID(r.Context(), orderID)
	if err != nil {
		http.Error(w, "Error getting order by ID", http.StatusInternalServerError)
		slog.Error("failed to get order by id", "error", err)
		return
	}

	//TODO: make list of verified roles and range by it
	if order.UserID != requestingUserID && auth.Role(requestingUserRole) != auth.RoleAdmin && auth.Role(requestingUserRole) != auth.RoleManager {
		http.Error(w, "Forbidden: You do not have permisions to view this order", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		http.Error(w, "User ID is missing", http.StatusBadRequest)
		return
	}

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
		slog.Error("failed to call to restaurants-service via gRPC", "error", err)
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
		UserID:       userID,
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
		slog.Error("failed to create order", "error", err)
		http.Error(w, "Error creating order", http.StatusInternalServerError)
		return
	}

	eventBody, err := json.Marshal(order)
	if err != nil {
		slog.Error("failed to marshal order for Kafka event", "error", err)
	} else {
		h.producer.Produce(r.Context(), messaging.OrderCreatedTopic, []byte(order.ID), eventBody)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}
