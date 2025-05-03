package service

import (
	"errors"
	"time"

	"fin_service/internal/integration"
	"fin_service/internal/integration/cbr"
	"fin_service/internal/models"
	"fin_service/internal/repository"
	"fin_service/utils"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type CreditService struct {
	AccountRepo *repository.AccountRepository
	CreditRepo  *repository.CreditRepository
	UserRepo    *repository.UserRepository
	EmailSender integration.EmailSender
}

func NewCreditService(accountRepo *repository.AccountRepository, creditRepo *repository.CreditRepository, userRepo *repository.UserRepository, emailSender integration.EmailSender) *CreditService {
	return &CreditService{
		AccountRepo: accountRepo,
		CreditRepo:  creditRepo,
		UserRepo:    userRepo,
		EmailSender: emailSender,
	}
}

func (s *CreditService) CreateCredit(userID, accountID string, amount float64, termMonths int) error {
	account, err := s.AccountRepo.GetAccountByID(accountID)
	if err != nil || account.UserID != userID {
		log.Warn("Unauthorized credit request")
		return errors.New("account not found or unauthorized")
	}

	rate, err := cbr.GetCentralBankRate()
	if err != nil {
		log.WithError(err).Error("Failed to fetch key rate")
		return errors.New("unable to fetch key rate")
	}
	finalRate := rate // уже с +5% маржой

	credit := &models.Credit{
		ID:           uuid.New().String(),
		UserID:       userID,
		AccountID:    accountID,
		Amount:       amount,
		TermMonths:   termMonths,
		InterestRate: finalRate,
		CreatedAt:    time.Now(),
	}

	payment := utils.CalculateAnnuityPayment(amount, finalRate, termMonths)

	var schedule []*models.PaymentSchedule
	for i := 1; i <= termMonths; i++ {
		schedule = append(schedule, &models.PaymentSchedule{
			ID:       uuid.New().String(),
			CreditID: credit.ID,
			DueDate:  credit.CreatedAt.AddDate(0, i, 0),
			Amount:   payment,
			Paid:     false,
		})
	}

	if err := s.CreditRepo.SaveCredit(credit); err != nil {
		return err
	}
	if err := s.CreditRepo.SaveSchedule(schedule); err != nil {
		return err
	}

	if email, err := s.UserRepo.FindEmailByUserID(userID); err == nil {
		_ = s.EmailSender.SendCreditIssuedEmail(email)
	}

	log.WithFields(log.Fields{
		"user_id": userID,
		"account": accountID,
		"amount":  amount,
		"term":    termMonths,
		"rate":    finalRate,
		"monthly": payment,
	}).Info("Credit successfully created")

	return nil
}

func (s *CreditService) GetPaymentSchedule(userID, creditID string) ([]*models.PaymentSchedule, error) {
	ownerID, err := s.CreditRepo.GetCreditOwner(creditID)
	if err != nil || ownerID != userID {
		return nil, errors.New("unauthorized access to credit")
	}
	return s.CreditRepo.GetScheduleByCreditID(creditID)
}
