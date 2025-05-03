package models

import "time"

type Credit struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	AccountID    string    `json:"account_id"`
	Amount       float64   `json:"amount"`
	TermMonths   int       `json:"term_months"`
	InterestRate float64   `json:"interest_rate"` // годовая ставка, например 12.5
	CreatedAt    time.Time `json:"created_at"`
}

type PaymentSchedule struct {
	ID       string     `json:"id"`
	CreditID string     `json:"credit_id"`
	DueDate  time.Time  `json:"due_date"`
	Amount   float64    `json:"amount"`
	Paid     bool       `json:"paid"`
	PaidAt   *time.Time `json:"paid_at,omitempty"`
}
