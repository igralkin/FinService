package handler

import (
	"encoding/json"
	"net/http"

	"fin_service/internal/service"

	log "github.com/sirupsen/logrus"
)

type CardRequest struct {
	AccountID string `json:"account_id"`
}

type CardResponse struct {
	CardID    string `json:"card_id"`
	CreatedAt string `json:"created_at"`
}

type CardHandler struct {
	CardService *service.CardService
}

func NewCardHandler(service *service.CardService) *CardHandler {
	return &CardHandler{CardService: service}
}

func (h *CardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.URL.Path != "/cards" {
		http.NotFound(w, r)
		return
	}

	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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
