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
		card.ExpiryMonthEnc,
		card.ExpiryYearEnc,
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

func (r *CardRepository) GetCardsByUserID(userID string) ([]*models.Card, error) {
	query := `
		SELECT id, user_id, account_id, number_enc, expiry_month_enc, expiry_year_enc,
		       cvv_hash, hmac, created_at
		FROM cards
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.DB.Query(query, userID)
	if err != nil {
		log.WithError(err).Error("Failed to query cards by user ID")
		return nil, err
	}
	defer rows.Close()

	var cards []*models.Card
	for rows.Next() {
		var card models.Card
		if err := rows.Scan(
			&card.ID,
			&card.UserID,
			&card.AccountID,
			&card.NumberEnc,
			&card.ExpiryMonthEnc,
			&card.ExpiryYearEnc,
			&card.CVVHash,
			&card.HMAC,
			&card.CreatedAt,
		); err != nil {
			log.WithError(err).Error("Failed to scan card row")
			continue
		}
		cards = append(cards, &card)
	}
	return cards, nil
}

func (r *CardRepository) GetCardByID(cardID string) (*models.Card, error) {
	query := `
		SELECT id, user_id, account_id, number_enc, expiry_month_enc, expiry_year_enc,
		       cvv_hash, hmac, created_at
		FROM cards
		WHERE id = $1
	`

	var card models.Card
	err := r.DB.QueryRow(query, cardID).Scan(
		&card.ID,
		&card.UserID,
		&card.AccountID,
		&card.NumberEnc,
		&card.ExpiryMonthEnc,
		&card.ExpiryYearEnc,
		&card.CVVHash,
		&card.HMAC,
		&card.CreatedAt,
	)

	if err != nil {
		log.WithError(err).Error("Failed to fetch card by ID")
		return nil, errors.New("card not found")
	}

	return &card, nil
}
