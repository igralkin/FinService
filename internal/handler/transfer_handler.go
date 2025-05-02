package handler

import (
	"encoding/json"
	"net/http"

	"fin_service/internal/service"

	log "github.com/sirupsen/logrus"
)

type TransferRequest struct {
	FromAccountID string  `json:"from_account_id"`
	ToAccountID   string  `json:"to_account_id"`
	Amount        float64 `json:"amount"`
}

type TransferResponse struct {
	Message string `json:"message"`
}

type TransferHandler struct {
	TransferService *service.TransferService
}

func NewTransferHandler(service *service.TransferService) *TransferHandler {
	return &TransferHandler{TransferService: service}
}

func (h *TransferHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.TransferService.Transfer(userID, req.FromAccountID, req.ToAccountID, req.Amount); err != nil {
		log.WithError(err).Warn("Transfer failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := TransferResponse{Message: "Transfer successful"}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
