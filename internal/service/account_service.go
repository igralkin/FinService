package service

import (
	"errors"
	"fin_service/internal/integration"
	"fin_service/internal/models"
	"fin_service/internal/repository"
	"time"

	log "github.com/sirupsen/logrus"
)

type AccountService struct {
	Repo        *repository.AccountRepository
	UserRepo    *repository.UserRepository
	EmailSender integration.EmailSender
	TxRepo      *repository.TransactionRepository
}

func NewAccountService(
	accountRepo *repository.AccountRepository,
	userRepo *repository.UserRepository,
	emailSender integration.EmailSender,
	txRepo *repository.TransactionRepository,
) *AccountService {
	return &AccountService{
		Repo:        accountRepo,
		UserRepo:    userRepo,
		EmailSender: emailSender,
		TxRepo:      txRepo,
	}
}

// CreateAccountForUser creates a new account and sends notification email.
func (s *AccountService) CreateAccountForUser(userID string) (string, error) {
	log.WithField("user_id", userID).Info("Creating new account")

	account := models.NewAccount(userID)
	accountID, err := s.Repo.CreateAccount(account)
	if err != nil {
		log.WithError(err).Error("Account creation failed")
		return "", err
	}

	email, err := s.UserRepo.FindEmailByUserID(userID)
	if err != nil {
		log.WithError(err).Warn("Could not find email for user after bank account creation")
		// Не блокируем создание счёта, просто не отправим письмо
		return accountID, nil
	}

	if err := s.EmailSender.SendAccountCreatedEmail(email); err != nil {
		log.WithError(err).Warn("Failed to send bank account creation email")
		// Тоже не блокируем
	}

	log.WithFields(log.Fields{
		"account_id": accountID,
		"user_id":    userID,
	}).Info("Bank account created and notification email sent")

	return accountID, nil
}

// GetBalance проверяет владельца и возвращает баланс счёта.
func (s *AccountService) GetBalance(userID, accountID string) (float64, error) {
	log.WithFields(log.Fields{
		"user_id":    userID,
		"account_id": accountID,
	}).Info("Request to get account balance")

	account, err := s.Repo.GetAccountByID(accountID)
	if err != nil {
		return 0, err
	}

	if account.UserID != userID {
		log.Warn("User tried to access another user's account")
		return 0, errors.New("unauthorized access")
	}

	return account.Balance, nil
}

// ListAccounts возвращает список всех счетов пользователя.
func (s *AccountService) ListAccounts(userID string) ([]*models.Account, error) {
	log.WithField("user_id", userID).Info("Listing user accounts")
	return s.Repo.GetAccountsByUserID(userID)
}

// Deposit adds funds to the specified account after validation.
func (s *AccountService) Deposit(userID, accountID string, amount float64) (float64, error) {
	log.WithFields(log.Fields{
		"user_id":    userID,
		"account_id": accountID,
		"amount":     amount,
	}).Info("Attempting bank account deposit")

	if amount <= 0 {
		return 0, errors.New("amount must be greater than zero")
	}

	account, err := s.Repo.GetAccountByID(accountID)
	if err != nil {
		return 0, err
	}

	if account.UserID != userID {
		log.Warn("Unauthorized deposit attempt")
		return 0, errors.New("unauthorized access")
	}

	newBalance, err := s.Repo.AddToBalance(accountID, amount)
	_ = s.TxRepo.LogTransaction(accountID, "deposit", amount, "User-initiated deposit")
	if err != nil {
		return 0, err
	}

	email, err := s.UserRepo.FindEmailByUserID(userID)
	if err == nil {
		_ = s.EmailSender.SendDepositEmail(email, amount)
	}

	return newBalance, nil
}

// Withdraw снимает средства со счёта после проверки.
func (s *AccountService) Withdraw(userID, accountID string, amount float64) (float64, error) {
	log.WithFields(log.Fields{
		"user_id":    userID,
		"account_id": accountID,
		"amount":     amount,
	}).Info("Attempting bank account withdrawal")

	if amount <= 0 {
		return 0, errors.New("amount must be greater than zero")
	}

	account, err := s.Repo.GetAccountByID(accountID)
	if err != nil {
		return 0, err
	}

	if account.UserID != userID {
		log.Warn("Unauthorized withdrawal attempt")
		return 0, errors.New("unauthorized access")
	}

	newBalance, err := s.Repo.SubtractFromBalance(accountID, amount)
	_ = s.TxRepo.LogTransaction(accountID, "withdraw", amount, "User-initiated withdrawal")
	if err != nil {
		return 0, err
	}

	email, err := s.UserRepo.FindEmailByUserID(userID)
	if err == nil {
		_ = s.EmailSender.SendWithdrawalEmail(email, amount)
	}

	return newBalance, nil
}

func (s *AccountService) GetTransactionHistory(userID, accountID string) ([]*models.Transaction, error) {
	log.WithFields(log.Fields{
		"user_id":    userID,
		"account_id": accountID,
	}).Info("Request to get transaction history")

	account, err := s.Repo.GetAccountByID(accountID)
	if err != nil {
		return nil, err
	}

	if account.UserID != userID {
		log.Warn("User tried to access another user's transaction history")
		return nil, errors.New("unauthorized access")
	}

	return s.TxRepo.GetTransactionsByAccountID(accountID)
}

func (s *AccountService) GetAnalytics(userID, accountID, monthStr string) (float64, float64, error) {
	account, err := s.Repo.GetAccountByID(accountID)
	if err != nil {
		return 0, 0, err
	}
	if account.UserID != userID {
		return 0, 0, errors.New("unauthorized access")
	}

	parsed, err := time.Parse("2006-01", monthStr)
	if err != nil {
		return 0, 0, errors.New("invalid month format, use YYYY-MM")
	}

	income, expense, err := s.TxRepo.GetAnalytics(accountID, parsed.Year(), int(parsed.Month()))
	if err != nil {
		return 0, 0, err
	}

	return income, expense, nil
}
