package handler

import (
	"encoding/json"
	"fin_service/internal/service"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

type CreditRequest struct {
	AccountID  string  `json:"account_id"`
	Amount     float64 `json:"amount"`
	TermMonths int     `json:"term_months"`
}

type CreditHandler struct {
	CreditService *service.CreditService
}

func NewCreditHandler(cs *service.CreditService) *CreditHandler {
	return &CreditHandler{CreditService: cs}
}

func (h *CreditHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodPost && r.URL.Path == "/credits" {
		var req CreditRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 || req.TermMonths <= 0 {
			http.Error(w, "Invalid credit request", http.StatusBadRequest)
			return
		}

		if err := h.CreditService.CreateCredit(userID, req.AccountID, req.Amount, req.TermMonths); err != nil {
			log.WithError(err).Warn("Failed to issue credit")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"message": "Credit created successfully"}`))
		return
	}

	if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/credits/") && strings.HasSuffix(r.URL.Path, "/schedule") {
		h.getSchedule(w, r, userID)
		return
	}

	http.NotFound(w, r)
}

func (h *CreditHandler) getSchedule(w http.ResponseWriter, r *http.Request, userID string) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	creditID := parts[2]

	schedule, err := h.CreditService.GetPaymentSchedule(userID, creditID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"schedule": schedule,
	})
}
