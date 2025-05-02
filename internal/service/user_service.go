package service

import (
	"errors"

	"fin_service/internal/integration"
	"fin_service/internal/models"
	"fin_service/internal/repository"

	"golang.org/x/crypto/bcrypt"
	log "github.com/sirupsen/logrus"
)

type UserService struct {
	Repo       *repository.UserRepository
	EmailSender integration.EmailSender
}

// NewUserService creates a new UserService instance.
func NewUserService(repo *repository.UserRepository, sender integration.EmailSender) *UserService {
	return &UserService{
		Repo:        repo,
		EmailSender: sender,
	}
}

// Register validates and registers a user, then sends a welcome email.
func (s *UserService) Register(email, username, password string) error {
	log.WithField("email", email).Info("Starting user registration")

	// Basic validation
	if err := models.ValidateBasic(email, username, password); err != nil {
		log.WithError(err).Warn("Validation failed")
		return err
	}

	// Uniqueness checks
	if s.Repo.IsEmailTaken(email) {
		log.WithField("email", email).Warn("Email already taken")
		return errors.New("email is already registered")
	}

	if s.Repo.IsUsernameTaken(username) {
		log.WithField("username", username).Warn("Username already taken")
		return errors.New("username is already taken")
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.WithError(err).Error("Failed to hash password")
		return errors.New("internal error")
	}

	// Create and save user
	user := models.NewUser(email, username, string(hashed))
	if err := s.Repo.SaveUser(user); err != nil {
		log.WithError(err).Error("Failed to save user")
		return err
	}

	// Send welcome email
	if err := s.EmailSender.SendWelcomeEmail(email); err != nil {
		log.WithError(err).Error("Failed to send welcome email")
		return errors.New("user registered, but failed to send email")
	}

	log.WithField("email", email).Info("User registered and welcome email sent")
	return nil
}
