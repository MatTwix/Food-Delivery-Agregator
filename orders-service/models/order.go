package models

import (
	"database/sql"
	"time"
)

type Order struct {
	ID            string         `json:"id"`
	RestaurantID  string         `json:"restaurant_id"`
	UserID        string         `json:"user_id"`
	TotalPrice    float64        `json:"total_price"`
	Status        string         `json:"status"`
	RetryCount    int            `json:"retry_count"`
	MaxRetryCount int            `json:"max_retry_count"`
	NextRetryAt   time.Time      `json:"next_retry_at"`
	CourierID     sql.NullString `json:"courier_id,omitempty"`
	Items         []OrderItem    `json:"items"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type OrderItem struct {
	ID         string  `json:"id,omitempty"`
	OrderID    string  `json:"order_id,omitempty"`
	MenuItemID string  `json:"menu_item_id"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
}
