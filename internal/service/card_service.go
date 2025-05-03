package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"fin_service/internal/integration"
	"fin_service/internal/models"
	"fin_service/internal/repository"
	"fin_service/utils"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type CardService struct {
	AccountRepo *repository.AccountRepository
	CardRepo    *repository.CardRepository
	UserRepo    *repository.UserRepository
	TxRepo      *repository.TransactionRepository
	EmailSender integration.EmailSender
}

func NewCardService(
	accountRepo *repository.AccountRepository,
	cardRepo *repository.CardRepository,
	userRepo *repository.UserRepository,
	txRepo *repository.TransactionRepository,
	emailSender integration.EmailSender,
) *CardService {
	return &CardService{
		AccountRepo: accountRepo,
		CardRepo:    cardRepo,
		UserRepo:    userRepo,
		TxRepo:      txRepo,
		EmailSender: emailSender,
	}
}

func (s *CardService) GenerateCard(userID, accountID string) (*models.Card, error) {
	account, err := s.AccountRepo.GetAccountByID(accountID)
	if err != nil {
		return nil, errors.New("account not found")
	}
	if account.UserID != userID {
		return nil, errors.New("unauthorized account access")
	}

	cardNumber := utils.GenerateValidCardNumber()
	expiry := time.Now().AddDate(3, 0, 0)
	expiryMonth := int(expiry.Month())
	expiryYear := expiry.Year()
	cvv := fmt.Sprintf("%03d", rand.Intn(1000))

	numberEnc, err := utils.EncryptPGP(cardNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt card number: %w", err)
	}

	monthEnc, err := utils.EncryptPGP(fmt.Sprintf("%02d", expiryMonth))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt expiry month: %w", err)
	}

	yearEnc, err := utils.EncryptPGP(fmt.Sprintf("%d", expiryYear))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt expiry year: %w", err)
	}

	cvvHash, err := utils.HashCVV(cvv)
	if err != nil {
		return nil, fmt.Errorf("failed to hash CVV: %w", err)
	}

	hmacInput := cardNumber + fmt.Sprintf("%02d", expiryMonth) + fmt.Sprintf("%d", expiryYear)
	hmac, err := utils.GenerateHMAC(hmacInput)
	if err != nil {
		return nil, fmt.Errorf("failed to generate HMAC: %w", err)
	}

	card := &models.Card{
		ID:             uuid.New().String(),
		UserID:         userID,
		AccountID:      accountID,
		NumberEnc:      numberEnc,
		ExpiryMonthEnc: monthEnc,
		ExpiryYearEnc:  yearEnc,
		CVVHash:        cvvHash,
		HMAC:           hmac,
		CreatedAt:      time.Now(),
	}

	if err := s.CardRepo.SaveCard(card); err != nil {
		log.WithError(err).Error("Failed to save card")
		return nil, errors.New("could not generate card")
	}

	if email, err := s.UserRepo.FindEmailByUserID(userID); err == nil {
		_ = s.EmailSender.SendCardCreatedEmail(email)
	}

	log.WithFields(log.Fields{
		"user_id":    userID,
		"account_id": accountID,
	}).Info("Card successfully generated")

	return card, nil
}

func (s *CardService) GetUserCards(userID string) ([]*models.Card, error) {
	return s.CardRepo.GetCardsByUserID(userID)
}

func (s *CardService) PayWithCard(userID, cardID string, amount float64, description string) error {
	card, err := s.CardRepo.GetCardByID(cardID)
	if err != nil {
		return err
	}
	if card.UserID != userID {
		return errors.New("unauthorized card access")
	}

	monthStr, _ := utils.DecryptPGP(card.ExpiryMonthEnc)
	yearStr, _ := utils.DecryptPGP(card.ExpiryYearEnc)
	expiryTime, _ := time.Parse("2006-01", yearStr+"-"+monthStr)
	if time.Now().After(expiryTime.AddDate(0, 1, -1)) {
		return errors.New("card is expired")
	}

	if _, err := s.AccountRepo.SubtractFromBalance(card.AccountID, amount); err != nil {
		return err
	}

	if err := s.TxRepo.LogTransaction(card.AccountID, "withdraw", amount, "Card payment: "+description); err != nil {
		log.WithError(err).Error("Failed to log card payment transaction")
	}

	if email, err := s.UserRepo.FindEmailByUserID(userID); err == nil {
		_ = s.EmailSender.SendCardPaymentEmail(email)
	}

	log.WithFields(log.Fields{
		"user_id": userID,
		"card_id": card.ID,
		"amount":  amount,
		"account": card.AccountID,
	}).Info("Card payment processed")

	return nil
}
