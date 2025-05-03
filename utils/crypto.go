package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// GenerateHMAC creates HMAC-SHA256 from provided message and secret.
func GenerateHMAC(data string) (string, error) {
	secret := os.Getenv("HMAC_SECRET")
	if secret == "" {
		return "", errors.New("HMAC_SECRET not set in environment")
	}
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil)), nil
}

// HashCVV returns a bcrypt hash of the CVV.
func HashCVV(cvv string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
