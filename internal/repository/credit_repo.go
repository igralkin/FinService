package repository

import (
	"database/sql"
	"fin_service/internal/models"
	"time"
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

// Получить предстоящие платежи пользователя
func (r *CreditRepository) GetUpcomingPayments(userID string, until time.Time) ([]*models.PaymentSchedule, error) {
	query := `
		SELECT ps.id, ps.credit_id, ps.due_date, ps.amount, ps.paid, ps.paid_at
		FROM payment_schedules ps
		JOIN credits c ON ps.credit_id = c.id
		WHERE c.user_id = $1 AND ps.paid = false AND ps.due_date <= $2
		ORDER BY ps.due_date
	`

	rows, err := r.DB.Query(query, userID, until)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*models.PaymentSchedule
	for rows.Next() {
		var p models.PaymentSchedule
		err := rows.Scan(&p.ID, &p.CreditID, &p.DueDate, &p.Amount, &p.Paid, &p.PaidAt)
		if err != nil {
			continue
		}
		payments = append(payments, &p)
	}
	return payments, nil
}

func (r *CreditRepository) GetDuePayments(before time.Time) ([]*models.PaymentSchedule, error) {
	rows, err := r.DB.Query(`
		SELECT id, credit_id, due_date, amount, paid, paid_at
		FROM payment_schedules
		WHERE paid = false AND due_date <= $1
	`, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*models.PaymentSchedule
	for rows.Next() {
		var p models.PaymentSchedule
		err := rows.Scan(&p.ID, &p.CreditID, &p.DueDate, &p.Amount, &p.Paid, &p.PaidAt)
		if err != nil {
			continue
		}
		payments = append(payments, &p)
	}
	return payments, nil
}

func (r *CreditRepository) GetAccountIDByCreditID(creditID string) (string, error) {
	var accountID string
	err := r.DB.QueryRow(`SELECT account_id FROM credits WHERE id = $1`, creditID).Scan(&accountID)
	return accountID, err
}

func (r *CreditRepository) MarkPaymentAsPaid(paymentID string) error {
	_, err := r.DB.Exec(`
		UPDATE payment_schedules SET paid = true, paid_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, paymentID)
	return err
}
