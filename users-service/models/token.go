package models

import "time"

type Token struct {
	TokenID   string    `json:"token_id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
