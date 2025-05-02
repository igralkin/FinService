package models

import "time"

type Account struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

// NewAccount creates a new account instance before inserting into DB.
func NewAccount(userID string) *Account {
	return &Account{
		UserID:  userID,
		Balance: 0.0,
	}
}
