package models

import "time"

type Delivery struct {
	OrderID    string    `json:"order_id"`
	CourierID  string    `json:"courier_id"`
	AssignedAt time.Time `json:"assigned_at"`
}
