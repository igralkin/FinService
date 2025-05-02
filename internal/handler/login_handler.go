package handler

import (
	"encoding/json"
	"net/http"

	"fin_service/internal/service"
	log "github.com/sirupsen/logrus"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type LoginHandler struct {
	AuthService *service.AuthService
}

// NewLoginHandler creates a new handler for user login.
func NewLoginHandler(authService *service.AuthService) *LoginHandler {
	return &LoginHandler{
		AuthService: authService,
	}
}

// ServeHTTP handles POST /login requests.
func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Warn("Failed to decode login request")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	token, err := h.AuthService.Authenticate(req.Email, req.Password)
	if err != nil {
		log.WithError(err).Warn("Authentication failed")
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	resp := LoginResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.WithError(err).Error("Failed to write login response")
	}
}
