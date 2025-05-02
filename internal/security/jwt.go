package security

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
)

// ErrMissingJWTSecret is returned if JWT_SECRET environment variable is not set.
var ErrMissingJWTSecret = errors.New("JWT_SECRET environment variable is not set")

// GenerateJWTToken creates and signs a JWT token with a 24-hour expiration.
// It includes the userID as the subject of the token.
func GenerateJWTToken(userID string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Error("JWT_SECRET is not set in environment")
		return "", ErrMissingJWTSecret
	}

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		log.WithError(err).Error("Failed to sign JWT token")
		return "", err
	}

	log.WithField("userID", userID).Info("Generated JWT token successfully")
	return signedToken, nil
}
