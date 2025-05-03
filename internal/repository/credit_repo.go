package repository

import (
	"database/sql"
	"fin_service/internal/models"
)

type CreditRepository struct {
	DB *sql.DB
}

func NewCreditRepository(db *sql.DB) *CreditRepository {
	return &CreditRepository{DB: db}
}

func (r *CreditRepository) SaveCredit(c *models.Credit) error {
	_, err := r.DB.Exec(`
		INSERT INTO credits (id, user_id, account_id, amount, term_months, interest_rate, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		c.ID, c.UserID, c.AccountID, c.Amount, c.TermMonths, c.InterestRate, c.CreatedAt)
	return err
}

func (r *CreditRepository) SaveSchedule(payments []*models.PaymentSchedule) error {
	for _, p := range payments {
		_, err := r.DB.Exec(`
			INSERT INTO payment_schedules (id, credit_id, due_date, amount, paid, paid_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, p.ID, p.CreditID, p.DueDate, p.Amount, p.Paid, p.PaidAt)
		if err != nil {
			return err
		}
	}
	return nil
}
func (r *CreditRepository) GetScheduleByCreditID(creditID string) ([]*models.PaymentSchedule, error) {
	rows, err := r.DB.Query(`
		SELECT id, credit_id, due_date, amount, paid, paid_at
		FROM payment_schedules
		WHERE credit_id = $1
		ORDER BY due_date
	`, creditID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedule []*models.PaymentSchedule
	for rows.Next() {
		var p models.PaymentSchedule
		err := rows.Scan(&p.ID, &p.CreditID, &p.DueDate, &p.Amount, &p.Paid, &p.PaidAt)
		if err != nil {
			continue
		}
		schedule = append(schedule, &p)
	}
	return schedule, nil
}

func (r *CreditRepository) GetCreditOwner(creditID string) (string, error) {
	var userID string
	err := r.DB.QueryRow(`SELECT user_id FROM credits WHERE id = $1`, creditID).Scan(&userID)
	if err != nil {
		return "", err
	}
	return userID, nil
}
