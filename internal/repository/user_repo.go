package repository

import (
	"database/sql"
	"errors"

	"fin_service/internal/models"

	log "github.com/sirupsen/logrus"
)

type UserRepository struct {
	DB *sql.DB
}

// NewUserRepository creates a new repository with a PostgreSQL connection.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// IsEmailTaken checks if the email already exists in the database.
func (r *UserRepository) IsEmailTaken(email string) bool {
	var exists bool
	err := r.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		log.WithError(err).Error("Failed to check email uniqueness")
		return true // fail-safe: assume taken
	}
	return exists
}

// IsUsernameTaken checks if the username already exists.
func (r *UserRepository) IsUsernameTaken(username string) bool {
	var exists bool
	err := r.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE username = $1)", username).Scan(&exists)
	if err != nil {
		log.WithError(err).Error("Failed to check username uniqueness")
		return true
	}
	return exists
}

// SaveUser inserts a new user into the database.
func (r *UserRepository) SaveUser(user *models.User) error {
	query := `
		INSERT INTO users (id, email, username, password_hash)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.DB.Exec(query, user.ID, user.Email, user.Username, user.PasswordHash)
	if err != nil {
		log.WithError(err).Error("Failed to insert user into database")
		return errors.New("failed to save user")
	}

	log.WithFields(log.Fields{
		"id":       user.ID,
		"email":    user.Email,
		"username": user.Username,
	}).Info("User saved to database successfully")

	return nil
}

// FindByEmail retrieves a user from the database by email.
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, username, password_hash
		FROM users
		WHERE email = $1
	`

	var user models.User
	err := r.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.WithField("email", email).Warn("User not found")
			return nil, errors.New("user not found")
		}
		log.WithError(err).Error("Failed to find user by email")
		return nil, errors.New("internal error")
	}

	log.WithField("email", email).Info("User found by email")
	return &user, nil
}

// FindEmailByUserID returns the email for a given user ID.
func (r *UserRepository) FindEmailByUserID(userID string) (string, error) {
	var email string
	query := `SELECT email FROM users WHERE id = $1`

	err := r.DB.QueryRow(query, userID).Scan(&email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.WithField("user_id", userID).Warn("User ID not found")
			return "", errors.New("user not found")
		}
		log.WithError(err).Error("Failed to retrieve email by user ID")
		return "", errors.New("internal error")
	}

	return email, nil
}
