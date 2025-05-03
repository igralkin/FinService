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
	EmailSender integration.EmailSender
}

func NewCardService(accountRepo *repository.AccountRepository, cardRepo *repository.CardRepository, userRepo *repository.UserRepository, emailSender integration.EmailSender) *CardService {
	return &CardService{
		AccountRepo: accountRepo,
		CardRepo:    cardRepo,
		UserRepo:    userRepo,
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

	numberEnc, _ := utils.EncryptPGP(cardNumber)
	monthEnc, _ := utils.EncryptPGP(fmt.Sprintf("%02d", expiryMonth))
	yearEnc, _ := utils.EncryptPGP(fmt.Sprintf("%d", expiryYear))
	cvvHash, _ := utils.HashCVV(cvv)

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

	email, err := s.UserRepo.FindEmailByUserID(userID)
	if err == nil {
		_ = s.EmailSender.SendCardCreatedEmail(email)
	}

	log.WithFields(log.Fields{
		"user_id":    userID,
		"account_id": accountID,
	}).Info("Card successfully generated")

	return card, nil
}
