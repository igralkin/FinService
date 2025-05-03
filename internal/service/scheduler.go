package service

import (
	"time"

	"fin_service/internal/repository"

	log "github.com/sirupsen/logrus"
)

type PaymentScheduler struct {
	CreditRepo  *repository.CreditRepository
	AccountRepo *repository.AccountRepository
	TxRepo      *repository.TransactionRepository
}

func NewPaymentScheduler(c *repository.CreditRepository, a *repository.AccountRepository, t *repository.TransactionRepository) *PaymentScheduler {
	return &PaymentScheduler{
		CreditRepo:  c,
		AccountRepo: a,
		TxRepo:      t,
	}
}

func (s *PaymentScheduler) RunEvery(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			s.ProcessDuePayments()
			<-ticker.C
		}
	}()
}

func (s *PaymentScheduler) ProcessDuePayments() {
	log.Info("Scheduler started: checking due credit payments")

	duePayments, err := s.CreditRepo.GetDuePayments(time.Now())
	if err != nil {
		log.WithError(err).Error("Scheduler: failed to load due payments")
		return
	}

	if len(duePayments) == 0 {
		log.Info("Scheduler: no due payments at this time")
		return
	}

	for _, p := range duePayments {
		accountID, err := s.CreditRepo.GetAccountIDByCreditID(p.CreditID)
		if err != nil {
			log.WithError(err).Warn("Scheduler: failed to find account for credit")
			continue
		}

		balance, err := s.AccountRepo.GetBalanceByID(accountID)
		if err != nil {
			continue
		}

		if balance >= p.Amount {
			_, _ = s.AccountRepo.SubtractFromBalance(accountID, p.Amount)
			_ = s.TxRepo.LogTransaction(accountID, "withdraw", p.Amount, "Auto payment")
			_ = s.CreditRepo.MarkPaymentAsPaid(p.ID)
			log.Infof("Auto payment: %.2f withdrawn from account %v", p.Amount, accountID)
		} else {
			penalty := p.Amount * 0.10
			_ = s.TxRepo.LogTransaction(accountID, "penalty", penalty, "Missed payment penalty")
			log.Warnf("Insufficient funds for payment %s: penalty %.2f applied", p.ID, penalty)
		}
	}
}
