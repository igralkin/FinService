package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"fin_service/internal/service"

	log "github.com/sirupsen/logrus"
)

type AccountResponse struct {
	AccountID string  `json:"account_id"`
	Balance   float64 `json:"balance"`
}

type AccountHandler struct {
	AccountService *service.AccountService
}

type DepositRequest struct {
	Amount float64 `json:"amount"`
}

type DepositResponse struct {
	Message    string  `json:"message"`
	NewBalance float64 `json:"new_balance"`
}

type WithdrawRequest struct {
	Amount float64 `json:"amount"`
}

type WithdrawResponse struct {
	Message    string  `json:"message"`
	NewBalance float64 `json:"new_balance"`
}

func NewAccountHandler(accountService *service.AccountService) *AccountHandler {
	return &AccountHandler{
		AccountService: accountService,
	}
}

func (h *AccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/accounts":
		h.createAccount(w, r, userID)
	case r.Method == http.MethodGet && r.URL.Path == "/accounts":
		h.listAccounts(w, r, userID)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/accounts/") && strings.HasSuffix(r.URL.Path, "/balance"):
		h.getBalance(w, r, userID)
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/deposit"):
		h.deposit(w, r, userID)
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/withdraw"):
		h.withdraw(w, r, userID)
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/transactions"):
		h.getTransactionHistory(w, r, userID)
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/analytics"):
		h.getAnalytics(w, r, userID)
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/predict"):
		h.predictBalance(w, r, userID)
	default:
		http.NotFound(w, r)
	}
}

func (h *AccountHandler) createAccount(w http.ResponseWriter, r *http.Request, userID string) {
	accountID, err := h.AccountService.CreateAccountForUser(userID)
	if err != nil {
		log.WithError(err).Warn("Account creation failed")
		http.Error(w, "Could not create account", http.StatusInternalServerError)
		return
	}

	resp := AccountResponse{AccountID: accountID}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *AccountHandler) listAccounts(w http.ResponseWriter, r *http.Request, userID string) {
	accounts, err := h.AccountService.ListAccounts(userID)
	if err != nil {
		log.WithError(err).Warn("Failed to list accounts")
		http.Error(w, "Could not list accounts", http.StatusInternalServerError)
		return
	}

	var response []AccountResponse
	for _, acc := range accounts {
		response = append(response, AccountResponse{
			AccountID: acc.ID,
			Balance:   acc.Balance,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h *AccountHandler) getBalance(w http.ResponseWriter, r *http.Request, userID string) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	accountID := parts[2]

	balance, err := h.AccountService.GetBalance(userID, accountID)
	if err != nil {
		log.WithError(err).Warn("Failed to get balance")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	resp := AccountResponse{
		AccountID: accountID,
		Balance:   balance,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *AccountHandler) deposit(w http.ResponseWriter, r *http.Request, userID string) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	accountID := parts[2]

	var req DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newBalance, err := h.AccountService.Deposit(userID, accountID, req.Amount)
	if err != nil {
		log.WithError(err).Warn("Deposit failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := DepositResponse{
		Message:    "Deposit successful",
		NewBalance: newBalance,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *AccountHandler) withdraw(w http.ResponseWriter, r *http.Request, userID string) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	accountID := parts[2]

	var req WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newBalance, err := h.AccountService.Withdraw(userID, accountID, req.Amount)
	if err != nil {
		log.WithError(err).Warn("Withdrawal failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := WithdrawResponse{
		Message:    "Withdrawal successful",
		NewBalance: newBalance,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *AccountHandler) getTransactionHistory(w http.ResponseWriter, r *http.Request, userID string) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	accountID := parts[2]

	transactions, err := h.AccountService.GetTransactionHistory(userID, accountID)
	if err != nil {
		log.WithError(err).Warn("Failed to get transaction history")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(transactions)
}

func (h *AccountHandler) getAnalytics(w http.ResponseWriter, r *http.Request, userID string) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	accountID := parts[2]

	month := r.URL.Query().Get("month")
	if month == "" {
		http.Error(w, "Missing 'month' query param", http.StatusBadRequest)
		return
	}

	income, expense, err := h.AccountService.GetAnalytics(userID, accountID, month)
	if err != nil {
		log.WithError(err).Warn("Failed to get analytics")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := map[string]interface{}{
		"account_id": accountID,
		"month":      month,
		"income":     income,
		"expense":    expense,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *AccountHandler) predictBalance(w http.ResponseWriter, r *http.Request, userID string) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	daysStr := r.URL.Query().Get("days")
	days, _ := strconv.Atoi(daysStr)
	if days < 1 || days > 365 {
		http.Error(w, "Days must be between 1 and 365", http.StatusBadRequest)
		return
	}

	current, predicted, err := h.AccountService.PredictBalance(userID, days)
	if err != nil {
		http.Error(w, "Failed to calculate forecast", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"days":              days,
		"current_balance":   current,
		"predicted_balance": predicted,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
