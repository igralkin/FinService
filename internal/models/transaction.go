package models

import "time"

type Transaction struct {
	ID          string    `json:"id"`
	AccountID   string    `json:"account_id"`
	Operation   string    `json:"operation"`   // "deposit", "withdraw", etc.
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
