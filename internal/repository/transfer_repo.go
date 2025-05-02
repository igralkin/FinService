package repository

import (
	"database/sql"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type TransferRepository struct {
	DB      *sql.DB
	TxRepo  *TransactionRepository
}

func NewTransferRepository(db *sql.DB, txRepo *TransactionRepository) *TransferRepository {
	return &TransferRepository{
		DB:     db,
		TxRepo: txRepo,
	}
}

// TransferFunds safely moves amount from one account to another within a transaction.
func (r *TransferRepository) TransferFunds(fromID, toID string, amount float64) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	// Проверка баланса и вычитание
	var fromBalance float64
	err = tx.QueryRow(`SELECT balance FROM accounts WHERE id = $1 FOR UPDATE`, fromID).Scan(&fromBalance)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to get source balance: %w", err)
	}
	if fromBalance < amount {
		_ = tx.Rollback()
		return errors.New("insufficient funds")
	}

	// Обновление баланса отправителя
	_, err = tx.Exec(`UPDATE accounts SET balance = balance - $1 WHERE id = $2`, amount, fromID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to debit sender: %w", err)
	}

	// Обновление баланса получателя
	_, err = tx.Exec(`UPDATE accounts SET balance = balance + $1 WHERE id = $2`, amount, toID)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to credit receiver: %w", err)
	}

	// Логируем транзакции
	if err := r.logTxInContext(tx, fromID, "transfer_out", amount, fmt.Sprintf("Transfer to %s", toID)); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := r.logTxInContext(tx, toID, "transfer_in", amount, fmt.Sprintf("Transfer from %s", fromID)); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transfer: %w", err)
	}

	log.WithFields(log.Fields{
		"from":   fromID,
		"to":     toID,
		"amount": amount,
	}).Info("Transfer completed successfully")

	return nil
}

// logTxInContext writes to the transactions table using the provided tx.
func (r *TransferRepository) logTxInContext(tx *sql.Tx, accountID, op string, amount float64, desc string) error {
	query := `
		INSERT INTO transactions (account_id, operation, amount, description)
		VALUES ($1, $2, $3, $4)
	`
	_, err := tx.Exec(query, accountID, op, amount, desc)
	if err != nil {
		log.WithError(err).Error("Failed to log transfer transaction")
		return fmt.Errorf("failed to log transaction: %w", err)
	}
	return nil
}
