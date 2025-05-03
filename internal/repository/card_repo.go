package repository

import (
	"database/sql"
	"errors"

	"fin_service/internal/models"
	log "github.com/sirupsen/logrus"
)

type CardRepository struct {
	DB *sql.DB
}

func NewCardRepository(db *sql.DB) *CardRepository {
	return &CardRepository{DB: db}
}

// SaveCard inserts a new virtual card into the database.
func (r *CardRepository) SaveCard(card *models.Card) error {
	query := `
		INSERT INTO cards (
			id, user_id, account_id,
			number_enc, expiry_month_enc, expiry_year_enc,
			cvv_hash, hmac, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.DB.Exec(
		query,
		card.ID,
		card.UserID,
		card.AccountID,
		card.NumberEnc,
		card.ExpiryMonth,  // временно, позже зашифруем
		card.ExpiryYear,   // временно, позже зашифруем
		card.CVVHash,
		card.HMAC,
		card.CreatedAt,
	)

	if err != nil {
		log.WithError(err).Error("Failed to insert card into database")
		return errors.New("failed to save card")
	}

	log.WithField("card_id", card.ID).Info("Card saved to database")
	return nil
}
