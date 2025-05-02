package service

import (
	"errors"

	"fin_service/internal/integration"
	"fin_service/internal/repository"

	log "github.com/sirupsen/logrus"
)

type TransferService struct {
	TransferRepo *repository.TransferRepository
	AccountRepo  *repository.AccountRepository
	UserRepo     *repository.UserRepository
	EmailSender  integration.EmailSender
}

func NewTransferService(
	transferRepo *repository.TransferRepository,
	accountRepo *repository.AccountRepository,
	userRepo *repository.UserRepository,
	emailSender integration.EmailSender,
) *TransferService {
	return &TransferService{
		TransferRepo: transferRepo,
		AccountRepo:  accountRepo,
		UserRepo:     userRepo,
		EmailSender:  emailSender,
	}
}

func (s *TransferService) Transfer(userID, fromID, toID string, amount float64) error {
	log.WithFields(log.Fields{
		"user_id": userID,
		"from":    fromID,
		"to":      toID,
		"amount":  amount,
	}).Info("Initiating bank account transfer")

	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	fromAccount, err := s.AccountRepo.GetAccountByID(fromID)
	if err != nil {
		return err
	}
	if fromAccount.UserID != userID {
		return errors.New("unauthorized access to from_account")
	}

	if fromID == toID {
		return errors.New("cannot transfer to the same account")
	}

	if err := s.TransferRepo.TransferFunds(fromID, toID, amount); err != nil {
		return err
	}

	// Отправка писем (не критично — ошибки не прерывают)
	if fromEmail, err := s.UserRepo.FindEmailByUserID(fromAccount.UserID); err == nil {
		_ = s.EmailSender.SendTransferSentEmail(fromEmail, amount, toID)
	}

	toAccount, err := s.AccountRepo.GetAccountByID(toID)
	if err == nil {
		if toEmail, err := s.UserRepo.FindEmailByUserID(toAccount.UserID); err == nil {
			_ = s.EmailSender.SendTransferReceivedEmail(toEmail, amount, fromID)
		}
	}

	return nil
}
