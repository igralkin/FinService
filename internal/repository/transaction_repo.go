package repository

import (
	"database/sql"
	"errors"
	"fin_service/internal/models"

	log "github.com/sirupsen/logrus"
)

type TransactionRepository struct {
	DB *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{DB: db}
}

// LogTransaction inserts a new transaction record.
func (r *TransactionRepository) LogTransaction(accountID, operation string, amount float64, description string) error {
	query := `
		INSERT INTO transactions (account_id, operation, amount, description)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.DB.Exec(query, accountID, operation, amount, description)
	if err != nil {
		log.WithError(err).Error("Failed to log transaction")
		return errors.New("failed to record transaction")
	}

	log.WithFields(log.Fields{
		"account_id": accountID,
		"operation":  operation,
		"amount":     amount,
	}).Info("Transaction logged")

	return nil
}

func (r *TransactionRepository) GetTransactionsByAccountID(accountID string) ([]*models.Transaction, error) {
	query := `
		SELECT id, account_id, operation, amount, description, created_at
		FROM transactions
		WHERE account_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.DB.Query(query, accountID)
	if err != nil {
		log.WithError(err).Error("Failed to query transaction history")
		return nil, errors.New("could not retrieve transactions")
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err := rows.Scan(
			&tx.ID,
			&tx.AccountID,
			&tx.Operation,
			&tx.Amount,
			&tx.Description,
			&tx.CreatedAt,
		); err != nil {
			log.WithError(err).Error("Failed to scan transaction row")
			continue
		}
		transactions = append(transactions, &tx)
	}

	return transactions, nil
}

func (r *TransactionRepository) GetAnalytics(accountID string, year int, month int) (income float64, expense float64, err error) {
	query := `
		SELECT operation, SUM(amount)
		FROM transactions
		WHERE account_id = $1
		AND EXTRACT(YEAR FROM created_at) = $2
		AND EXTRACT(MONTH FROM created_at) = $3
		GROUP BY operation
	`

	rows, err := r.DB.Query(query, accountID, year, month)
	if err != nil {
		log.WithError(err).Error("Failed to get analytics from transactions")
		return 0, 0, errors.New("could not retrieve analytics")
	}
	defer rows.Close()

	for rows.Next() {
		var operation string
		var total float64
		if err := rows.Scan(&operation, &total); err != nil {
			log.WithError(err).Error("Failed to scan analytics row")
			continue
		}

		switch operation {
		case "deposit", "transfer_in":
			income += total
		case "withdraw", "transfer_out":
			expense += total
		}
	}

	return income, expense, nil
}
