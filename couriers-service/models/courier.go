package models

import "time"

type Courier struct {
	ID        string    `json:"id"`
	Name      string    `json:"string"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
