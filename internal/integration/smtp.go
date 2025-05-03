package integration

import (
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

type EmailSender interface {
	SendWelcomeEmail(to string) error
	SendLoginNotificationEmail(to string) error
	SendAccountCreatedEmail(to string) error
	SendDepositEmail(to string, amount float64) error
	SendWithdrawalEmail(to string, amount float64) error
	SendTransferSentEmail(to string, amount float64, toAccountID string) error
	SendTransferReceivedEmail(to string, amount float64, fromAccountID string) error
	SendCardCreatedEmail(to string) error
}

type SMTPClient struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
	SSL      bool
}

// NewSMTPClientFromEnv initializes SMTP client using environment variables.
func NewSMTPClientFromEnv() (*SMTPClient, error) {
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP_PORT: %w", err)
	}

	return &SMTPClient{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     port,
		User:     os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
		SSL:      true,
	}, nil
}

// SendWelcomeEmail sends a welcome message to the given email address.
func (s *SMTPClient) SendWelcomeEmail(to string) error {
	log.WithField("to", to).Info("Sending welcome email")

	m := gomail.NewMessage()
	m.SetHeader("From", s.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Welcome to BankGo!")
	m.SetBody("text/plain", "Hello! Your account has been successfully created.")

	d := gomail.NewDialer(s.Host, s.Port, s.User, s.Password)
	d.SSL = s.SSL

	if err := d.DialAndSend(m); err != nil {
		log.WithError(err).Error("Failed to send welcome email")
		return err
	}

	log.WithField("to", to).Info("Welcome email sent successfully")
	return nil
}

// SendLoginNotificationEmail sends a login confirmation email to the user.
func (s *SMTPClient) SendLoginNotificationEmail(to string) error {
	log.WithField("to", to).Info("Sending login notification email")

	m := gomail.NewMessage()
	m.SetHeader("From", s.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Login to BankGo")
	m.SetBody("text/plain", "You have successfully logged in to your account.")

	d := gomail.NewDialer(s.Host, s.Port, s.User, s.Password)
	d.SSL = s.SSL

	if err := d.DialAndSend(m); err != nil {
		log.WithError(err).Error("Failed to send login email")
		return err
	}

	log.WithField("to", to).Info("Login email sent successfully")
	return nil
}

// SendAccountCreatedEmail sends a notification about account creation.
func (s *SMTPClient) SendAccountCreatedEmail(to string) error {
	log.WithField("to", to).Info("Sending bank account creation email")

	m := gomail.NewMessage()
	m.SetHeader("From", s.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "New Bank Account Created")
	m.SetBody("text/plain", "A new bank account has been successfully created.")

	d := gomail.NewDialer(s.Host, s.Port, s.User, s.Password)
	d.SSL = s.SSL

	if err := d.DialAndSend(m); err != nil {
		log.WithError(err).Error("Failed to send bank account creation email")
		return err
	}

	log.WithField("to", to).Info("Bank account creation email sent successfully")
	return nil
}

// SendDepositEmail notifies the user of a successful deposit.
func (s *SMTPClient) SendDepositEmail(to string, amount float64) error {
	log.WithFields(log.Fields{
		"to":     to,
		"amount": amount,
	}).Info("Sending bank account deposit email")

	m := gomail.NewMessage()
	m.SetHeader("From", s.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Bank account deposit")
	m.SetBody("text/plain", fmt.Sprintf("Your account has been credited with %.2f RUB.", amount))

	d := gomail.NewDialer(s.Host, s.Port, s.User, s.Password)
	d.SSL = s.SSL

	if err := d.DialAndSend(m); err != nil {
		log.WithError(err).Error("Failed to send deposit email")
		return err
	}

	log.WithField("to", to).Info("Bank account deposit email sent successfully")
	return nil
}

// SendWithdrawalEmail отправляет уведомление о снятии средств.
func (s *SMTPClient) SendWithdrawalEmail(to string, amount float64) error {
	log.WithFields(log.Fields{
		"to":     to,
		"amount": amount,
	}).Info("Sending bank account withdrawal email")

	m := gomail.NewMessage()
	m.SetHeader("From", s.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Bank account withdrawal")
	m.SetBody("text/plain", fmt.Sprintf("An amount of %.2f RUB has been withdrawn from your account.", amount))

	d := gomail.NewDialer(s.Host, s.Port, s.User, s.Password)
	d.SSL = s.SSL

	if err := d.DialAndSend(m); err != nil {
		log.WithError(err).Error("Failed to send withdrawal email")
		return err
	}

	log.WithField("to", to).Info("Bank account withdrawal email sent successfully")
	return nil
}

// SendTransferSentEmail notifies sender about a successful transfer.
func (s *SMTPClient) SendTransferSentEmail(to string, amount float64, toAccountID string) error {
	log.WithFields(log.Fields{
		"to":            to,
		"amount":        amount,
		"to_account_id": toAccountID,
	}).Info("Sending transfer sent email")

	m := gomail.NewMessage()
	m.SetHeader("From", s.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Bank account transfer sent")
	m.SetBody("text/plain", fmt.Sprintf("You have sent %.2f RUB to account %s.", amount, toAccountID))

	d := gomail.NewDialer(s.Host, s.Port, s.User, s.Password)
	d.SSL = s.SSL

	if err := d.DialAndSend(m); err != nil {
		log.WithError(err).Error("Failed to send transfer sent email")
		return err
	}

	return nil
}

// SendTransferReceivedEmail notifies recipient about received transfer.
func (s *SMTPClient) SendTransferReceivedEmail(to string, amount float64, fromAccountID string) error {
	log.WithFields(log.Fields{
		"to":              to,
		"amount":          amount,
		"from_account_id": fromAccountID,
	}).Info("Sending transfer received email")

	m := gomail.NewMessage()
	m.SetHeader("From", s.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Bank account transfer received")
	m.SetBody("text/plain", fmt.Sprintf("You have received %.2f RUB from account %s.", amount, fromAccountID))

	d := gomail.NewDialer(s.Host, s.Port, s.User, s.Password)
	d.SSL = s.SSL

	if err := d.DialAndSend(m); err != nil {
		log.WithError(err).Error("Failed to send transfer received email")
		return err
	}

	return nil
}

// SendCardCreatedEmail sends a notification when a new card is created.
func (s *SMTPClient) SendCardCreatedEmail(to string) error {
	log.WithField("to", to).Info("Sending card creation email")

	m := gomail.NewMessage()
	m.SetHeader("From", s.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "New Virtual Card Created")
	m.SetBody("text/plain", "Your new virtual bank card has been successfully created.")

	d := gomail.NewDialer(s.Host, s.Port, s.User, s.Password)
	d.SSL = s.SSL

	if err := d.DialAndSend(m); err != nil {
		log.WithError(err).Error("Failed to send card creation email")
		return err
	}

	log.WithField("to", to).Info("Card creation email sent successfully")
	return nil
}
