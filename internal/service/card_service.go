package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"fin_service/internal/models"
	"fin_service/internal/repository"
	"fin_service/utils"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type CardService struct {
	AccountRepo *repository.AccountRepository
	CardRepo    *repository.CardRepository
}

func NewCardService(accountRepo *repository.AccountRepository, cardRepo *repository.CardRepository) *CardService {
	return &CardService{
		AccountRepo: accountRepo,
		CardRepo:    cardRepo,
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

	// Генерация данных
	cardNumber := utils.GenerateValidCardNumber()
	expiry := time.Now().AddDate(3, 0, 0)
	expiryMonth := int(expiry.Month())
	expiryYear := expiry.Year()
	cvv := fmt.Sprintf("%03d", rand.Intn(1000))

	// Шифрование / хеширование
	numberEnc, _ := utils.EncryptPGP(cardNumber)
	monthEnc, _ := utils.EncryptPGP(fmt.Sprintf("%02d", expiryMonth))
	yearEnc, _ := utils.EncryptPGP(fmt.Sprintf("%d", expiryYear))
	cvvHash, _ := utils.HashCVV(cvv)

	// HMAC по открытым данным
	hmacInput := cardNumber + fmt.Sprintf("%02d", expiryMonth) + fmt.Sprintf("%d", expiryYear)
	hmac, _ := utils.GenerateHMAC(hmacInput)

	card := &models.Card{
		ID:          uuid.New().String(),
		UserID:      userID,
		AccountID:   accountID,
		NumberEnc:   numberEnc,
		ExpiryMonth: expiryMonth,
		ExpiryYear:  expiryYear,
		CVVHash:     cvvHash,
		HMAC:        hmac,
		CreatedAt:   time.Now(),
	}

	if err := s.CardRepo.SaveCard(card); err != nil {
		log.WithError(err).Error("Failed to save card")
		return nil, errors.New("could not generate card")
	}

	log.WithFields(log.Fields{
		"user_id":    userID,
		"account_id": accountID,
	}).Info("Card successfully generated")

	return card, nil
}
