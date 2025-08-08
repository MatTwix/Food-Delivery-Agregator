package models

import "time"

type MenuItem struct {
	ID           string    `json:"id"`
	RestaurantID string    `json:"restaurant_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Price        int       `json:"price"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
