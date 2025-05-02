package models

import (
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"` // Не возвращается клиенту
}

// NewUser creates a new User with a generated UUID.
func NewUser(email, username, passwordHash string) *User {
	return &User{
		ID:           uuid.NewString(),
		Email:        strings.TrimSpace(email),
		Username:     strings.TrimSpace(username),
		PasswordHash: passwordHash,
	}
}

// ValidateBasic checks basic email, username, and password rules.
// The plainPassword must be at least 8 characters.
func ValidateBasic(email, username, plainPassword string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("invalid email format")
	}

	if utf8.RuneCountInString(username) < 3 {
		return errors.New("username must be at least 3 characters")
	}

	if utf8.RuneCountInString(plainPassword) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	return nil
}
