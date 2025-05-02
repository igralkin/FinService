package service

import (
	"errors"
	"os"
	"time"

	"fin_service/internal/integration"
	"fin_service/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Repo        *repository.UserRepository
	EmailSender integration.EmailSender
}

func NewAuthService(repo *repository.UserRepository, sender integration.EmailSender) *AuthService {
	return &AuthService{
		Repo:        repo,
		EmailSender: sender,
	}
}

func (s *AuthService) Authenticate(email, password string) (string, error) {
	log.WithField("email", email).Info("Starting user authentication")

	user, err := s.Repo.FindByEmail(email)
	if err != nil {
		log.WithError(err).Warn("User not found")
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		log.WithError(err).Warn("Invalid password")
		return "", errors.New("invalid credentials")
	}

	// Send login confirmation email
	if err := s.EmailSender.SendLoginNotificationEmail(email); err != nil {
		log.WithError(err).Error("Failed to send login notification email")
		// Не блокируем вход, просто логируем
	}

	// Generate JWT
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Error("JWT_SECRET not set")
		return "", errors.New("internal error")
	}

	claims := jwt.RegisteredClaims{
		Subject:   user.ID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		log.WithError(err).Error("Failed to sign JWT")
		return "", errors.New("internal error")
	}

	log.WithField("email", email).Info("User authenticated, JWT issued")
	return tokenStr, nil
}
