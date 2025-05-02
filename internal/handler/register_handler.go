package handler

import (
	"encoding/json"
	"net/http"

	"fin_service/internal/service"

	log "github.com/sirupsen/logrus"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Message string `json:"message"`
}

type RegisterHandler struct {
	UserService *service.UserService
}

// NewRegisterHandler creates a new HTTP handler for user registration.
func NewRegisterHandler(userService *service.UserService) *RegisterHandler {
	return &RegisterHandler{
		UserService: userService,
	}
}

// ServeHTTP processes POST /register requests.
func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Warn("Failed to decode request body")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := h.UserService.Register(req.Email, req.Username, req.Password); err != nil {
		log.WithError(err).Warn("User registration failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := RegisterResponse{
		Message: "Registration successful, welcome email sent.",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.WithError(err).Error("Failed to write response")
	}
}
