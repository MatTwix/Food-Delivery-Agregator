package models

import "time"

type Order struct {
	ID           string      `json:"id"`
	RestaurantID string      `json:"restaurant_id"`
	UserID       string      `json:"user_id,omitempty"`
	TotalPrice   float64     `json:"total_price"`
	Status       string      `json:"status"`
	Items        []OrderItem `json:"items"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID         string  `json:"id"`
	OrderID    string  `json:"order_id"`
	MenuItemID string  `json:"menu_item_id"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
}
