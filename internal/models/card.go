package models

import "time"

type Card struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	AccountID   string    `json:"account_id"`
	NumberEnc   string    `json:"-"` // Зашифрованный номер (PGP)
	ExpiryMonth int       `json:"-"` // Зашифрованный (PGP)
	ExpiryYear  int       `json:"-"` // Зашифрованный (PGP)
	CVVHash     string    `json:"-"` // Хеш CVV (bcrypt)
	HMAC        string    `json:"-"` // HMAC от всех данных
	CreatedAt   time.Time `json:"created_at"`
}
