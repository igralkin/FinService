package repository

import (
	"database/sql"
	"errors"
	"fin_service/internal/models"

	log "github.com/sirupsen/logrus"
)

type AccountRepository struct {
	DB *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{DB: db}
}

// CreateAccount inserts a new account for the given user and returns its ID.
func (r *AccountRepository) CreateAccount(account *models.Account) (string, error) {
	query := `
		INSERT INTO accounts (user_id, balance)
		VALUES ($1, $2)
		RETURNING id
	`

	var accountID string
	err := r.DB.QueryRow(query, account.UserID, account.Balance).Scan(&accountID)
	if err != nil {
		log.WithError(err).Error("Failed to create account")
		return "", errors.New("could not create account")
	}

	log.WithFields(log.Fields{
		"account_id": accountID,
		"user_id":    account.UserID,
	}).Info("Account created successfully")

	return accountID, nil
}

// GetAccountByID возвращает счёт по ID.
func (r *AccountRepository) GetAccountByID(accountID string) (*models.Account, error) {
	query := `
		SELECT id, user_id, balance, created_at
		FROM accounts
		WHERE id = $1
	`

	var acc models.Account
	err := r.DB.QueryRow(query, accountID).Scan(
		&acc.ID,
		&acc.UserID,
		&acc.Balance,
		&acc.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.WithField("account_id", accountID).Warn("Account not found")
			return nil, errors.New("account not found")
		}
		log.WithError(err).Error("Failed to retrieve account")
		return nil, errors.New("internal error")
	}

	return &acc, nil
}

// GetAccountsByUserID возвращает список счетов пользователя.
func (r *AccountRepository) GetAccountsByUserID(userID string) ([]*models.Account, error) {
	query := `
		SELECT id, user_id, balance, created_at
		FROM accounts
		WHERE user_id = $1
	`

	rows, err := r.DB.Query(query, userID)
	if err != nil {
		log.WithError(err).Error("Failed to query accounts by user ID")
		return nil, errors.New("internal error")
	}
	defer rows.Close()

	var accounts []*models.Account
	for rows.Next() {
		var acc models.Account
		if err := rows.Scan(&acc.ID, &acc.UserID, &acc.Balance, &acc.CreatedAt); err != nil {
			log.WithError(err).Error("Failed to scan account row")
			continue
		}
		accounts = append(accounts, &acc)
	}

	return accounts, nil
}

// AddToBalance увеличивает баланс счёта на указанную сумму.
func (r *AccountRepository) AddToBalance(accountID string, amount float64) (float64, error) {
	query := `
		UPDATE accounts
		SET balance = balance + $1
		WHERE id = $2
		RETURNING balance
	`

	var newBalance float64
	err := r.DB.QueryRow(query, amount, accountID).Scan(&newBalance)
	if err != nil {
		log.WithError(err).Error("Failed to update account balance (deposit)")
		return 0, errors.New("could not deposit to account")
	}

	log.WithFields(log.Fields{
		"account_id":  accountID,
		"amount":      amount,
		"new_balance": newBalance,
	}).Info("Bank account deposit successful")

	return newBalance, nil
}

// SubtractFromBalance уменьшает баланс счёта на указанную сумму.
func (r *AccountRepository) SubtractFromBalance(accountID string, amount float64) (float64, error) {
	query := `
		UPDATE accounts
		SET balance = balance - $1
		WHERE id = $2 AND balance >= $1
		RETURNING balance
	`

	var newBalance float64
	err := r.DB.QueryRow(query, amount, accountID).Scan(&newBalance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.WithFields(log.Fields{
				"account_id": accountID,
				"amount":     amount,
			}).Warn("Insufficient funds or account not found")
			return 0, errors.New("insufficient funds or account not found")
		}
		log.WithError(err).Error("Failed to subtract from account balance")
		return 0, errors.New("could not withdraw from account")
	}

	log.WithFields(log.Fields{
		"account_id":  accountID,
		"amount":      amount,
		"new_balance": newBalance,
	}).Info("Bank account withdrawal successful")

	return newBalance, nil
}

// Получить сумму по всем счетам пользователя
func (r *AccountRepository) GetTotalBalanceByUserID(userID string) (float64, error) {
	var total float64
	err := r.DB.QueryRow(`
		SELECT COALESCE(SUM(balance), 0)
		FROM accounts
		WHERE user_id = $1
	`, userID).Scan(&total)
	return total, err
}

func (r *AccountRepository) GetBalanceByID(accountID string) (float64, error) {
	var balance float64
	err := r.DB.QueryRow(`
		SELECT balance
		FROM accounts
		WHERE id = $1
	`, accountID).Scan(&balance)
	if err != nil {
		log.WithError(err).Error("Failed to get account balance by ID")
		return 0, errors.New("account not found")
	}
	return balance, nil
}
