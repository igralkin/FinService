package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"fin_service/internal/service"
	"fin_service/utils"

	log "github.com/sirupsen/logrus"
)

type CardRequest struct {
	AccountID string `json:"account_id"`
}

type CardResponse struct {
	CardID    string `json:"card_id"`
	CreatedAt string `json:"created_at"`
}

type CardListResponse struct {
	CardID       string `json:"card_id"`
	AccountID    string `json:"account_id"`
	CreatedAt    string `json:"created_at"`
	MaskedNumber string `json:"masked_number"`
	Expiry       string `json:"expiry"`
}
type CardPaymentRequest struct {
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}
type CardHandler struct {
	CardService *service.CardService
}

func NewCardHandler(service *service.CardService) *CardHandler {
	return &CardHandler{CardService: service}
}

func (h *CardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodPost && r.URL.Path == "/cards" {
		h.createCard(w, r, userID)
		return
	} else if r.Method == http.MethodGet && r.URL.Path == "/cards" {
		h.listCards(w, r, userID)
		return
	} else if r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/cards/") && strings.HasSuffix(r.URL.Path, "/pay") {
		h.payWithCard(w, r, userID)
		return
	}

	http.NotFound(w, r)
}

func (h *CardHandler) createCard(w http.ResponseWriter, r *http.Request, userID string) {
	var req CardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.AccountID == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	card, err := h.CardService.GenerateCard(userID, req.AccountID)
	if err != nil {
		log.WithError(err).Warn("Card generation failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := CardResponse{
		CardID:    card.ID,
		CreatedAt: card.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *CardHandler) listCards(w http.ResponseWriter, r *http.Request, userID string) {
	cards, err := h.CardService.GetUserCards(userID)
	if err != nil {
		log.WithError(err).Warn("Failed to get cards")
		http.Error(w, "Failed to get cards", http.StatusInternalServerError)
		return
	}

	var resp []CardListResponse
	for _, c := range cards {
		last4 := utils.MaskCardNumber(c.NumberEnc)
		expMonth, _ := utils.DecryptPGP(c.ExpiryMonthEnc)
		expYear, _ := utils.DecryptPGP(c.ExpiryYearEnc)

		resp = append(resp, CardListResponse{
			CardID:       c.ID,
			AccountID:    c.AccountID,
			CreatedAt:    c.CreatedAt.Format("2006-01-02 15:04:05"),
			MaskedNumber: last4,
			Expiry:       expMonth + "/" + expYear[len(expYear)-2:],
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"cards": resp})
}

func (h *CardHandler) payWithCard(w http.ResponseWriter, r *http.Request, userID string) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	cardID := parts[2]

	var req CardPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 {
		http.Error(w, "Invalid payment request", http.StatusBadRequest)
		return
	}

	if err := h.CardService.PayWithCard(userID, cardID, req.Amount, req.Description); err != nil {
		log.WithError(err).Warn("Card payment failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message": "Payment successful"}`))
}